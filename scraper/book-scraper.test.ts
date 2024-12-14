import { kyoboScraper, type ScrapData } from "./book-scraper";
import { expectTypeOf } from 'expect-type';
import path from 'path';
import { describe, test as it, expect } from "bun:test";
import fsSync from "fs"

describe(('book scraper'), async () => {
    const specPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/spec.html"
    const searchPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/search.html"

    const scraper = new kyoboScraper();
    scraper.dataPath = "/Users/yangwoolee/repo/libra-data/scraper/temp/test"
    scraper.isbn = "9791163034735" // 점프 투 파이썬
    await scraper.initBrowser()

    it('should loadWebSpecPage ', async () => {
        const isloaded = await scraper.loadWebSpecPage();
        expect(isloaded).toBe(true)
    }, { timeout: 10_000 });


    it('should loadSpecPage ', async () => {
        const isloaded = await scraper.loadSpecPage();
        expect(isloaded).toBe(true)
    }, { timeout: 10_000 });

    it('should saveBookHtml ', async () => {
        await scraper.page.goto(specPath)
        await scraper.saveHtml();

        const localFilePath = path.join(scraper.dataPath, scraper.scraperName, scraper.isbn + ".html")
        expect(fsSync.existsSync(localFilePath)).toBe(true)
        // if (fsSync.existsSync(localFilePath)) {
        //     fsSync.rmSync(localFilePath)
        // }
    });
    it('should loadLocalSpecPage ', async () => {
        const isloaded = await scraper.loadLocalSpecPage();
        expect(isloaded).toBe(true)
    });
    it('should get image url', async () => {
        await scraper.page.goto(specPath)

        const src = await scraper.extractImageSrc()
    })

    it('should saveImage ', async () => {
        await scraper.page.goto(specPath)

        expect(await scraper.saveImage()).toBe(true)
    });

    it('should extract bookinfo', async () => {
        await scraper.page.goto(specPath)

        // console.log(await scraper.extractData())
        expectTypeOf(await scraper.extractData()).toMatchTypeOf({
            toc: "string",
            recommendation: "string",
            description: "string",
            source: "string",
            url: "string"
        })
    })

    it('should searchBook ', async () => {
        const url = await scraper.searchBook(searchPath);
        expect(url).toEqual("https://product.kyobobook.co.kr/detail/S000202532365")
    });

    it('should get Recommendation ', async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getRecommendation()).length).toBeGreaterThan(0)
        console.log("getRecommendation", (await scraper.getRecommendation()).length)
    });

    it('should get Toc ', async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getToc()).length).toBeGreaterThan(0)
        console.log("getToc", (await scraper.getToc()).length)
    });

    it('should get Description ', async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getDescription()).length).toBeGreaterThan(0)
        console.log("getDescription", (await scraper.getDescription()).length)
    });
})