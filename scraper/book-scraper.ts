import { chromium, type Page, type Browser, BrowserContext } from "playwright"
import path from "path"
import { fetch } from "bun"
import fsAsync from "fs/promises"
import pino, { type Logger } from "pino"
import pretty from "pino-pretty"
import { format } from "date-fns"

const SEARCH_URL =
    "https://search.kyobobook.co.kr/search?gbCode=TOT&target=total"

export async function initBrowser(
    headless: boolean = false
): Promise<BrowserContext> {
    const bravePath =
        "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"
    const browser = await chromium.launch({
        executablePath: bravePath,
        headless,
    })
    return await browser.newContext()
}

export async function loadTargets(): Promise<string[]> {
    const targetFile = Bun.file(
        path.join(__dirname, "data/target", "target.csv")
    )
    const targetText = await targetFile.text()
    var csvArr = targetText.split("\n")

    if (csvArr[0] !== "isbn,status") {
        const newCsvArr = addStatusColumn(csvArr)
        await Bun.write(targetFile, newCsvArr.join("\n"))
        csvArr = newCsvArr
    }
    csvArr.shift() // drop header
    return csvArr.reduce((acc, row) => {
        const cols = row.split(",")
        if (cols[1] === "N") {
            acc.push(cols[0])
        }
        return acc
    }, [] as string[])
}

function addStatusColumn(csvArr: string[]): string[] {
    csvArr.shift() // drop header
    return [
        "isbn,status",
        ...csvArr.map((row) => [row.replace(/^\"|\"$/g, ""), "N"].join(",")),
    ]
}

export async function scrapIsbns(
    isbns: string[],
    numWorker: number,
    headless: boolean = false
): Promise<ScrapData[]> {
    const ctx = await initBrowser(headless)
    ctx.setDefaultTimeout(20000)
    const chunck = Math.ceil(isbns.length / numWorker)
    const isbnsChunk = Array(numWorker)
        .fill("")
        .map((_, idx) => isbns.slice(idx * chunck, (idx + 1) * chunck))
        .filter((item) => item.length > 0)

    const _numWorker = Math.min(isbnsChunk.length, numWorker)
    const workers = await Promise.all(
        Array(_numWorker)
            .fill(null)
            .map(async () => new kyoboScraper(await ctx.newPage()))
    )

    const result = await Promise.all(
        workers.map((worker, idx) => worker.execAll(isbnsChunk[idx]))
    )
    await ctx.close()
    return result.flat(1).filter((x) => x != null)
}

export async function saveScrapResult(data: ScrapData[]): Promise<string[]> {
    const fileName = format(Date.now(), "yyyyMMdd-HHmmss")
    const filePath = path.join(__dirname, "data/kyobo", fileName + ".json")
    await Bun.write(filePath, JSON.stringify(data), { createPath: true })
    return data.map((d) => d.isbn)
}

export async function updateTargetResult(
    targetArr: string[],
    resultArr: string[]
) {
    const targetFile = Bun.file(
        path.join(__dirname, "data/target", "target.csv")
    )
    const targetText = await targetFile.text()
    var csvArr = targetText.split("\n")

    const resultIsbnSet = new Set(resultArr)
    const notFoundSet = new Set(targetArr.filter((t) => !resultIsbnSet.has(t)))

    const newCsvArr = csvArr.map((rowStr) => {
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

    await Bun.write(targetFile, newCsvArr.join("\n"))
}

export type ScrapData = {
    isbn: string
    toc: string
    recommendation: string
    description: string
    source: string
    url: string
}

interface BookScraper {
    exec(): Promise<ScrapData | null>
    loadSpecPage(): Promise<[Boolean, "local" | "web" | null]>
    loadLocalSpecPage(): Promise<boolean>
    loadWebSpecPage(): Promise<boolean>
    saveHtml(): void
    saveImage(): Promise<boolean>
    searchBook(searchURL: string): Promise<string>
    extractData(): Promise<ScrapData>
}

const LoggingFile = pino.destination({
    dest: path.join(__dirname, "scraplogger.log"),
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



export class kyoboScraper implements BookScraper {
    page!: Page
    isbn!: string
    scraperName = "kyobo"
    dataPath = "/Users/yangwoolee/repo/libra-data/scraper/temp"
    logger = LoggerInstance

    constructor(page: Page) {
        this.page = page
    }

    async execAll(isbns: string[] | undefined): Promise<(ScrapData | null)[]> {
        if (!isbns) return []
        const result: (ScrapData | null)[] = []
        for (const isbn of isbns) {
            this.isbn = isbn
            result.push(await this.exec())
        }
        return result
    }

    async exec(): Promise<ScrapData | null> {
        const [isSpecPageLoaded, loadType] = await this.loadSpecPage()
        if (!isSpecPageLoaded) {
            this.logger.error(`${this.isbn} not found!`)
            return null
        }
        loadType === "web" && await this.saveHtml()
        loadType === "web" && await this.saveImage()
        return await this.extractData()
    }

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
        const src =
            ((await loc.count()) && (await loc.first().getAttribute("src"))) ||
            ""
        this.logger.debug({ src }, "Extracted image source")

        return src
    }
    async saveImage(): Promise<boolean> {
        const src = await this.extractImageSrc()

        if (src) {
            const response = await fetch(src)
            const arrayBuffer = await response.arrayBuffer()
            const bookName = await this.getBookName()
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
    async searchBook(searchURL: string): Promise<string> {
        await this.page.goto(searchURL)
        await this.page.waitForLoadState("domcontentloaded")

        const selector = '//ul[@class="prod_list"]//a[@class="prod_link"]'
        const loc = this.page.locator(selector)
        this.logger.debug("possible books lengths : %d", await loc.count())
        const specUrl =
            (await loc.count()) > 0
                ? await loc.first().getAttribute("href")
                : ""
        this.logger.debug("selected books url : %s ", specUrl)
        return specUrl || ""
    }
    async extractData(): Promise<ScrapData> {
        const urlXpath = "meta[property='og:url']"
        const url = await this.page
            .locator(urlXpath)
            .first()
            .getAttribute("content")
        const toc = await this.getToc()
        const recommendation = await this.getRecommendation()
        const description = await this.getDescription()

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
        }
    }
    async getRecommendation(): Promise<string> {
        const recoXpathFirst =
            '//div[@class="product_detail_area book_publish_review"]'
        const loc = this.page.locator(recoXpathFirst)
        const recoXpathSec = '//p[@class="info_text"]'
        var loc2
        if ((await loc.count()) > 0) {
            loc2 = loc.locator(recoXpathSec)
        }
        return ((await loc2?.count()) && (await loc2?.innerText())) || ""
    }
    async getToc(): Promise<string> {
        const tocXpath = '//li[@class="book_contents_item"]'
        const loc = this.page.locator(tocXpath)
        return ((await loc.count()) && (await loc.innerText())) || ""
    }
    async getDescription(): Promise<string> {
        const descXpath = '//div[@class="intro_bottom"]'
        const loc = this.page.locator(descXpath)
        return ((await loc.count()) && (await loc.innerText())) || ""
    }
    async getBookName(): Promise<string> {
        const bookNameXpath = "//span[@class='prod_title']"
        const bookName = await this.page
            .locator(bookNameXpath)
            .first()
            .textContent()
        return bookName || ""
    }
}

export function randomNumber(min: number, max: number) {
    return Math.random() * (max - min) + min
}
