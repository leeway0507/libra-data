import { kyoboScraper, initBrowser, scrapBookData, updateTargetStatus } from "./book-scraper"
import { expectTypeOf } from "expect-type"
import path from "path"
import { describe, test as it, expect } from "bun:test"
import fsSync from "fs"

describe("multi pages scrap using book scraper", () => {
    it(
        "should scrap by multi pages",
        async () => {
            await scrapBookData(["9791163034735", "9791163034735"], 2)
        },
        { timeout: 10_000 }
    )

    it("should update target ", async () => {
        const target = ["1234", "5678"]
        const result = ["1234"]
        await updateTargetStatus(target, result)
    })
})

describe("book scraper", async () => {
    const browser = await initBrowser()

    const scraper = new kyoboScraper(await browser.newPage()) // 점프 투 파이썬
    scraper.dataPath = `${__dirname}/temp/test`
    scraper.isbn = "9791163034735"

    const searchPath = `file://${__dirname}/temp/test/kyobo/search.html`
    const specPath = `file://${__dirname}/temp/test/kyobo/${scraper.isbn}.html`

    it(
        "should load Web Spec Page ",
        async () => {
            const isloaded = await scraper.loadWebSpecPage()
            expect(isloaded).toBe(true)
        },
        { timeout: 10_000 }
    )

    it("should load Local Spec Page ", async () => {
        const isloaded = await scraper.loadLocalSpecPage()
        expect(isloaded).toBe(true)
    })

    it(
        "should load Spec Page ",
        async () => {
            const isloaded = await scraper.loadSpecPage()
            expect(isloaded[0]).toBe(true)
        },
        { timeout: 10_000 }
    )

    it("should save Book Html ", async () => {
        await scraper.page.goto(specPath)
        await scraper.saveHtml()

        const localFilePath = path.join(
            scraper.dataPath,
            scraper.scraperName,
            scraper.isbn + ".html"
        )
        expect(fsSync.existsSync(localFilePath)).toBe(true)
        // if (fsSync.existsSync(localFilePath)) {
        //     fsSync.rmSync(localFilePath)
        // }
    })
    it("should get image url", async () => {
        await scraper.page.goto(specPath)
        await scraper.extractImageSrc()
    })

    it("should save Image file ", async () => {
        await scraper.page.goto(specPath)

        expect(await scraper.saveImage()).toBe(true)
    })

    it("should extract book info", async () => {
        await scraper.page.goto(specPath)

        expectTypeOf(await scraper.extractDataFromSpecPage()).toMatchTypeOf({
            isbn: "string",
            toc: "string",
            recommendation: "string",
            description: "string",
            source: "string",
            url: "string",
            author: "string",
            title: "string",
            imageUrl: "string",
        })
    })

    it("should search Book ", async () => {
        const url = await scraper.searchBook(searchPath)
        expect(url).toEqual("https://product.kyobobook.co.kr/detail/S000202532365")
    })

    it("should get Recommendation ", async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getRecommendation()).length).toBeGreaterThan(0)
        console.log("getRecommendation", (await scraper.getRecommendation()).length)
    })

    it("should get Toc ", async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getToc()).length).toBeGreaterThan(0)
        console.log("getToc", (await scraper.getToc()).length)
    })

    it("should get Description ", async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getDescription()).length).toBeGreaterThan(0)
        console.log("getDescription", (await scraper.getDescription()).length)
    })
})
