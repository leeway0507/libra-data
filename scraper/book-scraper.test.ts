import { kyoboScraper, type ScrapData } from "./book-scraper";
import { expectTypeOf } from 'expect-type';
import path from 'path';
import fsAsync from "fs/promises"
import fsSync from "fs"

describe(('Consortium List SubScraper '), () => {
    const scraper = new kyoboScraper();
    scraper.dataPath = "/Users/yangwoolee/repo/libra-data/scraper/temp/test"
    scraper.isbn = "9791163034735" // 점프 투 파이썬

    test('should loadSpecPage ', () => {
        scraper.loadSpecPage();
        expect(scraper.page.url).toEqual("")
    });

    test('should loadLocalSpecPage ', () => {
        scraper.loadLocalSpecPage();
        expect(scraper.page.url).toEqual("")
    });

    test('should loadWebSpecPage ', () => {
        scraper.loadWebSpecPage();
        expect(scraper.page.url).toEqual("")
    });

    test('should saveBookHtml ', async () => {
        await scraper.initBrowser()
        const specPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/spec.html"
        await scraper.page.goto(specPath)

        await scraper.saveHtml();

        const localFilePath = path.join(scraper.dataPath, scraper.scraperName, scraper.isbn + ".html")
        expect(fsSync.existsSync(localFilePath)).toBe(true)
        if (fsSync.existsSync(localFilePath)) {
            fsSync.rmSync(localFilePath)
        }
    });

    test('should saveImage ', () => {
        expect(scraper.saveImage(Buffer.from("test"))).toEqual("ok")
    });

    test('should extract bookinfo', async () => {
        await scraper.initBrowser()
        const specPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/spec.html"
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

    test('should searchBook ', async () => {
        await scraper.initBrowser()
        const searchPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/search.html"
        const url = await scraper.searchBook(searchPath);

        expect(url).toEqual("https://product.kyobobook.co.kr/detail/S000202532365")
    });

    test('should getRecommendation ', () => {
        expect(scraper.getRecommendation()).toEqual("")
    });

    test('should getToc ', () => {
        expect(scraper.getToc()).toEqual("")
    });
})