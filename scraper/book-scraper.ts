import { chromium, type Page, type Locator } from 'playwright';
import fsAsync from 'fs/promises';
import path from 'path';

const DATA_PATH = "/Users/yangwoolee/repo/libra-data/scraper"
const SEARCH_URL = "https://search.kyobobook.co.kr/search?gbCode=TOT&target=total"

type ScrapData = {
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
    getRecommendation(): string
    getToc(): string
    getDescription(): string
}

export class kyoboScraper implements BookScraper {
    page!: Page;
    scraperName: string = "kyobo"
    isbn!: string


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
        const localFilePath = path.join(DATA_PATH, "temp", "html", this.scraperName, this.isbn + ".html")
        if (await fsAsync.exists(localFilePath)) {
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
    saveHtml(): void {

    }
    saveImage(byte: Buffer): "ok" | "fail" {
        return "ok"
    }
    async searchBook(searchURL: string): Promise<string> {
        await this.page.goto(searchURL)
        await this.page.waitForLoadState("domcontentloaded")

        const selector = '//ul[@class="prod_list"]//a[@class="prod_link"]';
        const loc = this.page.locator(selector);
        const specUrl = await loc.count() > 1 ? await loc.first().getAttribute("href") : ""
        console.log(await loc.count(), specUrl)
        return specUrl || ""
    }
    getRecommendation(): string {
        return ""
    }
    getToc(): string {
        return ""
    }
    getDescription(): string {
        return ""
    }

}

