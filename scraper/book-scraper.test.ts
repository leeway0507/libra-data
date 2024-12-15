import {
    kyoboScraper,
    initBrowser,
    scrapIsbns,
    updateTargetResult,
} from "./book-scraper"
import { expectTypeOf } from "expect-type"
import path from "path"
import { describe, test as it, expect } from "bun:test"
import fsSync from "fs"

describe("multi page scrap", () => {
    it(
        "should execute multi page ",
        async () => {
            await scrapIsbns(["9791163034735", "9791163034735"], 2)
        },
        { timeout: 10_000 }
    )

    it("should update target ", async () => {
        const target = ["1234", "5678"]
        const result = ["1234"]
        await updateTargetResult(target, result)
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
        "should loadWebSpecPage ",
        async () => {
            const isloaded = await scraper.loadWebSpecPage()
            expect(isloaded).toBe(true)
        },
        { timeout: 10_000 }
    )

    it(
        "should loadSpecPage ",
        async () => {
            const isloaded = await scraper.loadSpecPage()
            expect(isloaded).toBe(true)
        },
        { timeout: 10_000 }
    )

    it("should saveBookHtml ", async () => {
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
    it("should loadLocalSpecPage ", async () => {
        const isloaded = await scraper.loadLocalSpecPage()
        expect(isloaded).toBe(true)
    })
    it("should get image url", async () => {
        await scraper.page.goto(specPath)
        await scraper.extractImageSrc()
    })

    it("should saveImage ", async () => {
        await scraper.page.goto(specPath)

        expect(await scraper.saveImage()).toBe(true)
    })

    it("should extract bookinfo", async () => {
        await scraper.page.goto(specPath)

        // console.log(await scraper.extractData())
        expectTypeOf(await scraper.extractData()).toMatchTypeOf({
            isbn: "string",
            toc: "string",
            recommendation: "string",
            description: "string",
            source: "string",
            url: "string",
        })
    })

    it("should searchBook ", async () => {
        const url = await scraper.searchBook(searchPath)
        expect(url).toEqual(
            "https://product.kyobobook.co.kr/detail/S000202532365"
        )
    })

    it("should get Recommendation ", async () => {
        await scraper.page.goto(specPath)
        expect((await scraper.getRecommendation()).length).toBeGreaterThan(0)
        console.log(
            "getRecommendation",
            (await scraper.getRecommendation()).length
        )
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
