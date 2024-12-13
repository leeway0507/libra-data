import { chromium, type Page, type Locator } from 'playwright';
import fs from 'fs';
import path from 'path';

export type ProductProps = {
    brand: string
    productName: string
    productImgUrl: string
    productUrl: string
    currencyCode: string
    retailPrice: string
    salePrice: string
    isSale: boolean,
    korBrand?: string
    korProductName?: string
    productId: string
    gender?: string
    color?: string
    category?: string
    categorySpec?: string
};

export type ScrapResultProps = {
    status: string
    data: unknown[]
};

export type StoreBrandDataProps = {
    store_name: string
    brand_name: string
    brand_url: string
};

export type PageJobProps = {
    brandName: string
};
export type ListJobProps = {
    brandName: string
};

export interface OptionsProps {
    scrollCount: number
    maxPagination: number
    storeName: string
    scrapType: string
}

export const defaultOptions: OptionsProps = {
    scrollCount: 10,
    maxPagination: 10000,
    storeName: '',
    scrapType: '',
};

export interface SubScraperInterface {
    job: PageJobProps | ListJobProps | void
    options: OptionsProps
    execute: () => Promise<ScrapResultProps>
    executePageScrap: (scrapData: ProductProps[], pageNation: number) => Promise<ProductProps[]>
}

class SubScraper<T extends object> {
    page!: Page;
    jobImp: T;
    options: OptionsProps;

    constructor(options: OptionsProps = defaultOptions) {
        this.jobImp = Object();
        this.options = options;
    }
    public set job(job: T) {
        this.jobImp = job;
    }

    public get job(): T | void {
        const jobExist = Object.keys(this.jobImp).length;
        return jobExist ? this.jobImp : this.jobNotImplementedError();
    }

    jobNotImplementedError() {
        throw new Error('jobNotImplementedError');
    }

    async initBrowser(headless: boolean = false) {
        const bravePath = '/Applications/Brave Browser.app/Contents/MacOS/Brave Browser';
        const browser = await chromium.launch({
            executablePath: bravePath,
            headless,
        });
        this.page = await browser.newPage(); // Assign the new page to this.page
    }

    /* v8 ignore start */
    async execute(): Promise<ScrapResultProps> {
        const url = this.getUrl();
        await this.browserWait();
        await this.page.goto(url);
        await this.handleCookies();
        await this.browserWait();
        const scrapResult = await this.scrap();
        return scrapResult;
    }
    /* v8 ignore stop */

    getUrl(): string {
        throw new Error('Method "handleCookies" must be implemented');
    }

    getBrandData(): StoreBrandDataProps[] {
        throw new Error('Method "handleCookies" must be implemented');
    }

    getRandomInt(min: number, max: number) {
        return Math.floor(Math.random() * (max - min + 1)) + min;
    }

    async handleCookies() {
        throw new Error('Method "handleCookies" must be implemented');
    }

    async afterNextClick() {
        throw new Error('Method "afterNextClick" must be implemented');
    }

    /* v8 ignore start */
    async scrap(): Promise<ScrapResultProps> {
        const duplicatedData = await this.executePageScrap();
        const scrapData = this.dropDuplicate(duplicatedData);
        return { status: 'success', data: scrapData };
    }

    async executePageScrap(
        scrapData: ProductProps[] = [],
        pageNation: number = 0,
    ): Promise<ProductProps[]> {
        await this.page.waitForLoadState('domcontentloaded');
        await this.scrollYPage();
        const result = await this.extractCards();
        scrapData.push(...result);

        const pageNationCount = pageNation + 1;
        const nextPage = await this.hasNextPage();

        if (nextPage && pageNationCount < this.options.maxPagination) {
            await nextPage.click();
            await this.afterNextClick();
            const nextPageResult = await this.executePageScrap(scrapData, pageNationCount);
            return nextPageResult;
        }
        return scrapData;
    }
    /* v8 ignore stop */

    async scrollYPage(c: number = 0) {
        if (this.options.scrollCount === 0) return;

        const boxY = await this.hasNextPage().then(async (r) => r?.boundingBox().then((r2) => r2?.y));
        const randomNumber = boxY && boxY < 1000 ? boxY : this.getRandomInt(400, 1000);

        await this.page.evaluate((num: number) => {
            /* v8 ignore next */
            window.scrollBy(0, num);
        }, randomNumber);

        await this.browserWait();
        const sc = c + 1;

        if (boxY && boxY < 0) {
            return;
        }

        if (sc < this.options.scrollCount) {
            await this.scrollYPage(sc);
        }
    }

    async extractCards(): Promise<ProductProps[]> {
        throw new Error('Method "extractCards" must be implemented');
    }

    async hasNextPage(): Promise<Locator | null> {
        throw new Error('Method "hasNextPage" must be implemented');
    }

    dropDuplicate(scrapData: ProductProps[]): ProductProps[] {
        const jsonObject = scrapData.map((r) => JSON.stringify(r));
        const uniqueSet = new Set(jsonObject);
        return Array.from(uniqueSet).map((r) => JSON.parse(r));
    }

    /* v8 ignore start */
    failedResponse(): ScrapResultProps {
        return { status: 'fail', data: [] };
    }

    async browserWait(time: number = this.getRandomInt(200, 1000)) {
        await this.page.waitForTimeout(time);
    }

    async loadingWait() {
        console.log('watingNetworkLdle');
        await this.page.waitForLoadState('networkidle');
    }
    /* v8 ignore stop */

    async downloadImage(url: string, filePath: string) {
        this.createFolders(path.dirname(filePath));

        const res = this.page.waitForResponse(url);
        await this.page.goto(url, { waitUntil: 'domcontentloaded' });

        const buff = await res.then((r) => r.body());
        fs.writeFile(filePath, new Uint8Array(buff), (err) => {
            if (err) {
                console.error('Error writing file:', err);
            }
        });
    }
    createFolders(filePath: string) {
        // Check if the directory exists
        if (!fs.existsSync(filePath)) {
            // If it doesn't exist, create it
            fs.mkdirSync(filePath, { recursive: true });
            console.log('Directory created successfully.');
        } else {
            console.log('Directory already exists.');
        }
    }
}

export default SubScraper;