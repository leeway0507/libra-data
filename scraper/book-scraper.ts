import { chromium, type Page, type Browser, BrowserContext } from "playwright"
import path from "path"
import { fetch } from "bun"
import fsAsync from "fs/promises"
import pino, { type Logger } from "pino"
import pretty from "pino-pretty"
import { format } from "date-fns"

export async function initBrowser(headless: boolean = false): Promise<BrowserContext> {
    const bravePath = "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"
    const browser = await chromium.launch({
        executablePath: bravePath,
        headless,
    })
    return await browser.newContext()
}

// 수집할 도서 데이터 로드(target.csv)
export async function loadTargetData(): Promise<string[]> {
    const targetFile = Bun.file(path.join(__dirname, "data/target", "target.csv"))
    const targetText = await targetFile.text()
    var csvData = targetText.split("\n")

    if (csvData[0] !== "isbn,status") {
        const newCsvData = filterUnusedColumns(csvData)
        await Bun.write(targetFile, newCsvData.join("\n"))
        csvData = newCsvData
    }
    csvData.shift() // drop header
    return csvData.reduce((acc, row) => {
        const [isbn, isScraped] = row.split(",")
        if (isScraped === "N") {
            acc.push(isbn)
        }
        return acc
    }, [] as string[])
}

function filterUnusedColumns(csvArr: string[]): string[] {
    csvArr.shift() // drop header
    return ["isbn,status", ...csvArr.map((row) => [row.replace(/^\"|\"$/g, ""), "N"].join(","))]
}

export async function scrapBookData(
    isbns: string[],
    numWorker: number,
    headless: boolean = false
): Promise<ScrapData[]> {
    const ctx = await initBrowser(headless)
    ctx.setDefaultTimeout(20000)

    // 타겟 데이터를 워커 수에 맞게 분배
    const workerChunkSize = Math.ceil(isbns.length / numWorker)
    const chunkedTargetArr = Array(numWorker)
        .fill("")
        .map((_, idx) => isbns.slice(idx * workerChunkSize, (idx + 1) * workerChunkSize))
        .filter((item) => item.length > 0)

    const _numWorker = Math.min(chunkedTargetArr.length, numWorker)
    const workers = await Promise.all(
        Array(_numWorker)
            .fill(null)
            .map(async () => new kyoboScraper(await ctx.newPage()))
    )

    const result = await Promise.all(
        workers.map((worker, idx) => worker.execAll(chunkedTargetArr[idx]))
    )
    await ctx.close()
    return result.flat(1).filter((x) => x != null)
}

// 수집 결과를 도서 단위로 저장(ex 9791211111.json)
export async function saveResult(data: ScrapData[]): Promise<string[]> {
    const fileName = format(Date.now(), "yyyyMMdd-HHmmss")
    const filePath = path.join(__dirname, "data/kyobo", fileName + ".json")
    await Bun.write(filePath, JSON.stringify(data), { createPath: true })
    return data.map((d) => d.isbn)
}

// 수집 상태를 업데이트(성공 : Y, 도서를 찾을 수 없음 : notFound)
export async function updateTargetStatus(targetIsbns: string[], resultIsbns: string[]) {
    const resultIsbnSet = new Set(resultIsbns)
    const notFoundSet = new Set(targetIsbns.filter((t) => !resultIsbnSet.has(t)))

    const targetFile = Bun.file(path.join(__dirname, "data/target", "target.csv"))
    const targetText = await targetFile.text()
    var csvData = targetText.split("\n")
    const newCsvData = csvData.map((rowStr) => {
        const row = rowStr.split(",")

        if (resultIsbnSet.has(row[0])) {
            row.pop()
            row.push("Y")
        } else if (notFoundSet.has(row[0])) {
            row.pop()
            row.push("notFound")
        }
        return row.join(",")
    })

    await Bun.write(targetFile, newCsvData.join("\n"))
}

export type ScrapData = {
    title: string
    author: string
    isbn: string
    toc: string
    recommendation: string
    description: string
    source: string
    url: string
    imageUrl: string
}

interface BookScraper {
    execAll(isbns: string[] | undefined): Promise<(ScrapData | null)[]>
    exec(): Promise<ScrapData | null>
    loadSpecPage(): Promise<[Boolean, "local" | "web" | null]>
    loadLocalSpecPage(): Promise<boolean>
    loadWebSpecPage(): Promise<boolean>
    saveHtml(): void
    saveImage(): Promise<boolean>
    searchBook(searchURL: string): Promise<string>
    extractDataFromSpecPage(): Promise<ScrapData>
}

const LoggingFile = pino.destination({
    dest: path.join(__dirname, "kyobo_scraper.log"),
    append: "stack",
})

const LoggerInstance: Logger = pino(
    {
        timestamp: pino.stdTimeFunctions.isoTime,
        level: "debug",
    },
    pino.multistream([
        { level: "info", stream: pretty() },
        { level: "debug", stream: LoggingFile },
    ])
)

const SEARCH_URL = "https://search.kyobobook.co.kr/search?gbCode=TOT&target=total"

export class kyoboScraper implements BookScraper {
    page!: Page
    isbn!: string
    scraperName = "kyobo"
    dataPath = "/Users/yangwoolee/repo/libra-data/scraper/temp"
    logger = LoggerInstance

    constructor(page: Page) {
        this.page = page
    }

    // 모든 isbns에 대한 수집
    async execAll(isbns: string[] | undefined): Promise<(ScrapData | null)[]> {
        if (!isbns) return []
        const result: (ScrapData | null)[] = []
        for (const isbn of isbns) {
            this.isbn = isbn
            result.push(await this.exec())
        }
        return result
    }
    // 개별 isbns에 대한 수집
    async exec(): Promise<ScrapData | null> {
        const [isSpecPageLoaded, loadType] = await this.loadSpecPage()
        if (!isSpecPageLoaded) {
            this.logger.error(`${this.isbn} not found!`)
            return null
        }
        loadType === "web" && (await this.saveHtml())
        loadType === "web" && (await this.saveImage())
        return await this.extractDataFromSpecPage()
    }

    // 도서 상세 페이지를 브라우저에 로드.
    // 도서를 최초 수집 시 이미지와 html 파일 로컬에 저장
    // 신규 도서의 경우 web에서 도서 상세페이지 로드
    // 기 수집 도서를 불러 올 경우 local에서 도서 상세페이지 로드,
    async loadSpecPage(): Promise<[Boolean, "local" | "web" | null]> {
        const isLocalLoaded = await this.loadLocalSpecPage()
        if (isLocalLoaded) {
            this.logger.debug("local html loaded")
            return [true, "local"]
        }

        const isWebLoaded = await this.loadWebSpecPage()
        if (isWebLoaded) {
            this.logger.debug("web html loaded")
            await Bun.sleep(randomNumber(3, 5) * 1000)
            return [true, "web"]
        }
        return [false, null]
    }

    async loadLocalSpecPage(): Promise<boolean> {
        const localFilePath = path.join(
            this.dataPath,
            this.scraperName,
            "html",
            this.isbn + ".html"
        )
        if (await Bun.file(localFilePath).exists()) {
            this.logger.debug({
                localFilePath: path.join("file://", localFilePath),
            })
            await this.page.goto(path.join("file://", localFilePath))
            return true
        }
        this.logger.debug("localPath : not exist")
        return false
    }

    async loadWebSpecPage(): Promise<boolean> {
        const searchURL = new URL(SEARCH_URL)
        searchURL.searchParams.set("keyword", this.isbn)
        this.logger.debug({ searchURL }, "loadWebSpecPage")

        const specUrl = await this.searchBook(searchURL.toString())
        this.logger.debug({ specUrl }, "loadWebSpecPage")
        if (specUrl === "") return false
        await this.page.goto(specUrl)
        return true
    }

    // 해당 도서의 상세 페이지에 접근하기 위해 도서 검색 후 상세 페이지 url 수집
    async searchBook(searchURL: string): Promise<string> {
        await this.page.goto(searchURL)
        await this.page.waitForLoadState("domcontentloaded")

        const selector = '//ul[@class="prod_list"]//a[@class="prod_link"]'
        const loc = this.page.locator(selector)
        this.logger.debug("possible books lengths : %d", await loc.count())

        const specUrl = (await loc.count()) > 0 ? await loc.first().getAttribute("href") : ""
        this.logger.debug("selected books url : %s ", specUrl)
        return specUrl || ""
    }

    async saveHtml() {
        if (this.page.url().startsWith("file://")) return
        const localFilePath = path.join(
            this.dataPath,
            this.scraperName,
            "html",
            `${this.isbn}.html`
        )
        this.logger.debug({ localFilePath }, "saveHtml")
        await fsAsync.mkdir(path.dirname(localFilePath), { recursive: true })

        const isFileExist = await Bun.file(localFilePath).exists()
        this.logger.debug({ isFileExist }, "saveHtml")

        if (!isFileExist) {
            const html = await this.page.content()
            if (html === "<html><head></head><body></body></html>") {
                this.logger.error({ error: "html is empty" }, "saveHtml")
                return
            }
            return await Bun.write(Bun.file(localFilePath), html)
        }
    }

    async extractImageSrc(): Promise<string> {
        const imgXpath = '//div[contains(@class, "portrait_img_box")]/img'
        const loc = this.page.locator(imgXpath)
        const src = ((await loc.count()) && (await loc.first().getAttribute("src"))) || ""
        this.logger.debug({ src }, "Extracted image source")
        return src
    }

    async saveImage(): Promise<boolean> {
        const src = await this.extractImageSrc()

        if (src) {
            const response = await fetch(src)
            const arrayBuffer = await response.arrayBuffer()
            const bookName = await this.getTitle()
            const exName = path.extname(src)

            const imagePath = path.join(
                this.dataPath,
                this.scraperName,
                "image",
                `${this.isbn}-${bookName?.replaceAll("/", "_")}${exName}`
            )
            this.logger.debug({ imagePath }, "saveImage")
            await Bun.write(imagePath, arrayBuffer)
            return true
        }
        return false
    }

    // 상세 페이지 내 도서 정보 수집
    async extractDataFromSpecPage(): Promise<ScrapData> {
        const urlXpath = "meta[property='og:url']"
        const url = await this.page.locator(urlXpath).first().getAttribute("content")
        const toc = await this.getToc()
        const recommendation = await this.getRecommendation()
        const description = await this.getDescription()
        const tilte = await this.getTitle()
        const imageUrl = await this.extractImageSrc()

        this.logger.debug(
            {
                isbn: this.isbn,
                toc: toc.length,
                recommendation: recommendation.length,
                description: description.length,
                source: this.scraperName,
                url,
            },
            "extractData"
        )
        return {
            isbn: this.isbn,
            toc,
            recommendation,
            description,
            source: this.scraperName,
            url: url || "",
            imageUrl: imageUrl,
            author: "", //todo
            title: tilte,
        }
    }
    // 추천사 수집
    async getRecommendation(): Promise<string> {
        const recoXpathFirst = '//div[@class="product_detail_area book_publish_review"]'
        const loc = this.page.locator(recoXpathFirst)
        const recoXpathSec = '//p[@class="info_text"]'
        var loc2
        if ((await loc.count()) > 0) {
            loc2 = loc.locator(recoXpathSec)
        }
        return ((await loc2?.count()) && (await loc2?.innerText())) || ""
    }
    // 목차 수집
    async getToc(): Promise<string> {
        const tocXpath = '//li[@class="book_contents_item"]'
        const loc = this.page.locator(tocXpath)
        return ((await loc.count()) && (await loc.innerText())) || ""
    }

    // 소개사 수집
    async getDescription(): Promise<string> {
        const descXpath = '//div[@class="intro_bottom"]'
        const loc = this.page.locator(descXpath)
        return ((await loc.count()) && (await loc.innerText())) || ""
    }
    // 도서명 수집
    async getTitle(): Promise<string> {
        const bookNameXpath = "//span[@class='prod_title']"
        const bookName = await this.page.locator(bookNameXpath).first().textContent()
        return bookName || ""
    }
}

export function randomNumber(min: number, max: number) {
    return Math.random() * (max - min) + min
}
