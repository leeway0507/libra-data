import { chromium, type Page, type Locator } from 'playwright';
import fsAsync from 'fs/promises';
import fsSync from 'fs'
import path from 'path';


const SEARCH_URL = "https://search.kyobobook.co.kr/search?gbCode=TOT&target=total"

export type ScrapData = {
    toc: string
    recommendation: string
    description: string
    source: string
    url: string
}



interface BookScraper {
    exec(isbn: string): Promise<string | null>
    loadSpecPage(): void
    loadLocalSpecPage(): Promise<boolean>
    loadWebSpecPage(): Promise<boolean>
    saveHtml(): void
    saveImage(byte: Buffer): "ok" | "fail"
    searchBook(searchURL: string): Promise<string>
    extractData(): Promise<ScrapData>
}

export class kyoboScraper implements BookScraper {
    page!: Page;
    scraperName: string = "kyobo"
    isbn!: string
    dataPath: string = "/Users/yangwoolee/repo/libra-data/scraper/temp/html"


    async exec(isbn: string): Promise<string | null> {
        this.isbn = isbn
        const isSpecPageLoaded = await this.loadSpecPage()
        if (!isSpecPageLoaded) return `${this.isbn} not found!`

        return null
    }

    async initBrowser(headless: boolean = false) {
        const bravePath = '/Applications/Brave Browser.app/Contents/MacOS/Brave Browser';
        const browser = await chromium.launch({
            executablePath: bravePath,
            headless,
        });
        this.page = await browser.newPage(); // Assign the new page to this.page
    }

    async loadSpecPage() {
        const isLocalLoaded = await this.loadLocalSpecPage()
        if (isLocalLoaded) return true

        const isWebLoaded = await this.loadWebSpecPage()
        if (isWebLoaded) return true
        return false

    }
    async loadLocalSpecPage(): Promise<boolean> {
        const localFilePath = path.join(this.dataPath, this.scraperName, this.isbn + ".html")
        if (fsSync.existsSync(localFilePath)) {
            await this.page.goto(localFilePath)
            return true
        }
        return false


    }
    async loadWebSpecPage(): Promise<boolean> {
        const searchURL = new URL(SEARCH_URL)
        searchURL.searchParams.set("keyword", this.isbn)
        const specUrl = await this.searchBook(searchURL.toString())
        if (specUrl === "") return false
        await this.page.goto(specUrl)
        return true

    }
    async saveHtml() {
        const html = await this.page.content()
        const localFilePath = path.join(this.dataPath, this.scraperName, this.isbn + ".html")
        await fsAsync.mkdir(path.dirname(localFilePath), { recursive: true })
        const isFileExist = fsSync.existsSync(localFilePath)
        if (!isFileExist) return await fsAsync.writeFile(localFilePath, html, { flag: 'wx' })

    }
    saveImage(byte: Buffer): "ok" | "fail" {
        return "ok"
    }
    async searchBook(searchURL: string): Promise<string> {
        await this.page.goto(searchURL)
        await this.page.waitForLoadState("domcontentloaded")

        const selector = '//ul[@class="prod_list"]//a[@class="prod_link"]';
        const loc = this.page.locator(selector);
        const specUrl = await loc.count() > 0 ? await loc.first().getAttribute("href") : ""
        return specUrl || ""
    }
    async extractData(): Promise<ScrapData> {
        return {
            toc: await this.getToc(),
            recommendation: await this.getRecommendation(),
            description: await this.getDescription(),
            source: this.scraperName,
            url: this.page.url()
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
        console.log("count", await loc.count())
        return await loc.count() && await loc.textContent() || ""
    }
    async getDescription(): Promise<string> {
        const descXpath = '//div[@class="intro_bottom"]'
        const loc = this.page.locator(descXpath)
        return await loc.count() && await loc.textContent() || ""
    }

}

