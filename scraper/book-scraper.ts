import { chromium, type Page, type Locator } from 'playwright';
import path from 'path';
import { fetch } from 'bun';
import fsAsync from "fs/promises"
import pino, { type Logger } from "pino"
import pretty from "pino-pretty"


const CURRENT_PATH = "/Users/yangwoolee/repo/libra-data/scraper"
const SEARCH_URL = "https://search.kyobobook.co.kr/search?gbCode=TOT&target=total"

export type ScrapData = {
    toc: string
    recommendation: string
    description: string
    source: string
    url: string
}

const LoggingFile = pino.destination({
    dest: path.join(CURRENT_PATH, "scraplogger.log"),
    append: 'stack'
})

const streams = [
    { level: 'debug', stream: pretty() },
    { level: 'debug', stream: LoggingFile },
]

const LoggerInstance: Logger = pino(
    {
        timestamp: pino.stdTimeFunctions.isoTime,
        level: "debug",
    },
    pino.multistream(streams),
)


interface BookScraper {
    exec(isbn: string): Promise<ScrapData | null>
    loadSpecPage(): void
    loadLocalSpecPage(): Promise<boolean>
    loadWebSpecPage(): Promise<boolean>
    saveHtml(): void
    saveImage(): Promise<boolean>
    searchBook(searchURL: string): Promise<string>
    extractData(): Promise<ScrapData>
}

export class kyoboScraper implements BookScraper {
    page!: Page;
    scraperName = "kyobo"
    isbn!: string
    dataPath = "/Users/yangwoolee/repo/libra-data/scraper/temp/html"
    bravePath = '/Applications/Brave Browser.app/Contents/MacOS/Brave Browser';
    logger = LoggerInstance

    async exec(isbn: string): Promise<ScrapData | null> {
        this.isbn = isbn
        await this.initBrowser()
        const isSpecPageLoaded = await this.loadSpecPage()
        if (!isSpecPageLoaded) {
            this.logger.error(`${this.isbn} not found!`)
            return null
        }
        await this.saveHtml()
        await this.saveImage()
        return await this.extractData()
    }

    async initBrowser(headless: boolean = false) {
        if (!this.page) {
            this.logger.info("init browser")
            const browser = await chromium.launch({
                executablePath: this.bravePath,
                headless,
            });
            this.page = await browser.newPage(); // Assign the new page to this.page
        }
    }

    async loadSpecPage() {
        const isLocalLoaded = await this.loadLocalSpecPage()
        if (isLocalLoaded) {
            this.logger.debug("local html loaded")
            return true
        }

        const isWebLoaded = await this.loadWebSpecPage()
        if (isWebLoaded) {
            this.logger.debug("web html loaded")
            return true
        }
        return false

    }
    async loadLocalSpecPage(): Promise<boolean> {
        const localFilePath = path.join(this.dataPath, this.scraperName, this.isbn + ".html")
        if (await Bun.file(localFilePath).exists()) {
            this.logger.debug({ localFilePath: path.join("file://", localFilePath) })
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
        const pageUrl = this.page.url().split("https://")[1]
        const localFilePath = path.join(
            this.dataPath,
            this.scraperName,
            `${this.isbn}.html`
        )
        this.logger.debug({ localFilePath }, "saveHtml")
        await fsAsync.mkdir(path.dirname(localFilePath), { recursive: true })

        const isFileExist = await Bun.file(localFilePath).exists()
        this.logger.debug({ isFileExist }, "saveHtml")


        if (!isFileExist) {
            const html = await this.page.content()
            return await Bun.write(Bun.file(localFilePath), html)
        }

    }
    async extractImageSrc(): Promise<string> {
        const imgXpath = '//div[contains(@class, "portrait_img_box")]/img'
        const loc = this.page.locator(imgXpath)
        const src = await loc.count() && await loc.first().getAttribute("src") || ""
        this.logger.debug({ src }, "Extracted image source");


        return src
    }
    async saveImage(): Promise<boolean> {
        const src = await this.extractImageSrc()

        if (src) {
            const response = await fetch(src);
            const arrayBuffer = await response.arrayBuffer();

            const bookNameXpath = "//span[@class='prod_title']"
            const bookName = await this.page.locator(bookNameXpath).first().textContent()
            const exName = path.extname(src)

            const imagePath = path.join(
                this.dataPath,
                this.scraperName,
                "image",
                `${this.isbn}-${bookName}${exName}`
            )
            this.logger.debug({ imagePath }, "saveImage")
            await Bun.write(imagePath, arrayBuffer);
            return true
        }
        return false
    }
    async searchBook(searchURL: string): Promise<string> {
        await this.page.goto(searchURL)
        await this.page.waitForLoadState("domcontentloaded")

        const selector = '//ul[@class="prod_list"]//a[@class="prod_link"]';
        const loc = this.page.locator(selector);
        this.logger.debug("possible books lengths : %d", await loc.count())
        const specUrl = await loc.count() > 0 ? await loc.first().getAttribute("href") : ""
        this.logger.debug("selected books url : %s ", specUrl)
        return specUrl || ""
    }
    async extractData(): Promise<ScrapData> {
        const urlXpath = "meta[property='og:url']"
        const url = await this.page.locator(urlXpath).first().getAttribute("content")
        const toc = await this.getToc()
        const recommendation = await this.getRecommendation()
        const description = await this.getDescription()

        this.logger.debug({
            toc: toc.length,
            recommendation: recommendation.length,
            description: description.length,
            source: this.scraperName,
            url
        }, "extractData")
        return {
            toc,
            recommendation,
            description,
            source: this.scraperName,
            url: url || ""
        }

    }
    async getRecommendation(): Promise<string> {
        const recoXpathFirst = '//div[@class="product_detail_area book_publish_review"]'
        const loc = this.page.locator(recoXpathFirst)
        const recoXpathSec = '//p[@class="info_text"]'
        var loc2
        if (await loc.count() > 0) {
            loc2 = loc.locator(recoXpathSec)
        }
        return await loc2?.count() && await loc2?.textContent() || ""
    }
    async getToc(): Promise<string> {
        const tocXpath = '//li[@class="book_contents_item"]'
        const loc = this.page.locator(tocXpath)
        return await loc.count() && await loc.textContent() || ""
    }
    async getDescription(): Promise<string> {
        const descXpath = '//div[@class="intro_bottom"]'
        const loc = this.page.locator(descXpath)
        return await loc.count() && await loc.textContent() || ""
    }

}

