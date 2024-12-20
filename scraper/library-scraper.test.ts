import { LibScraper } from "./library-scraper"
import { initBrowser } from "./book-scraper"
import { describe, test as it, expect } from "bun:test"
import fs from "fs"
import path from "path"

process.env.DATA_PATH = "/Users/yangwoolee/repo/libra-data/data/test"

describe("library scraper", async () => {
    const ctx = await initBrowser()
    const libScraperInstance = new LibScraper(ctx)
    const newPage = await ctx.newPage()
    it(
        "should select location ",
        async () => {
            const isSelected = await libScraperInstance.selectLocation(newPage)
            expect(isSelected).toBe(true)
        },
        { timeout: 10_000 }
    )
    it(
        "should select LibType ",
        async () => {
            await newPage.goto("https://www.data4library.kr/openDataL")
            await newPage.waitForLoadState("domcontentloaded")
            const isSelected = await libScraperInstance.selectLibType(newPage)
            expect(isSelected).toBe(true)
        },
        { timeout: 10_000 }
    )
    it("should test rest", () => {
        Array(10)
            .fill("")
            .map((k, v) => {
                const value = v + 1
                console.log(value, (value % 5) - 1)
            })
    })
    it("x", () => {
        console.log(10 % 5)
        console.log(10 % 5)
    })
    it(
        "should move to target page",
        async () => {
            const targetPage = 7
            await newPage.goto("https://www.data4library.kr/openDataL")
            await newPage.waitForLoadState("domcontentloaded")

            await libScraperInstance.moveToTargetPagination(newPage, targetPage)

            const pageNationXPath = "//a[@class='page']"
            const loc = newPage.locator(pageNationXPath)
            const idx = (targetPage % 5) - 1
            const title = await loc.nth(idx === -1 ? 4 : idx).getAttribute("title")
            expect(title).toBe("현재 페이지")
        },
        { timeout: 10_000 }
    )
    it(
        "should get lib data and filter if files exist",
        async () => {
            await newPage.goto("https://www.data4library.kr/openDataL")
            await newPage.waitForLoadState("networkidle")
            const libArr = await libScraperInstance.getLibList(newPage)
            console.log(libArr)
        },
        { timeout: 10_000 }
    )
    it(
        "should filter if files exist",
        async () => {
            const libArrMock = [
                {
                    libName: "가락몰도서관",
                    uploadDate: "2024-12-01",
                    nth: 0,
                },
                {
                    libName: "KB국민은행과 함께하는 나무 작은도서관",
                    uploadDate: "2024-12-01",
                    nth: 1,
                },
            ]
            const libFilteredMock = libScraperInstance.exctractCandidate(libArrMock)
            expect(libFilteredMock.length).toBe(1)

            await libScraperInstance.moveToDownloadPage(newPage, libArrMock[0])
            expect(newPage.url()).toBe("https://www.data4library.kr/openDataV")
            newPage.waitForTimeout(1000)
        },
        { timeout: 10_000 }
    )
    it(
        "should move to download page",
        async () => {
            await newPage.goto("https://www.data4library.kr/openDataL")
            await newPage.waitForLoadState("domcontentloaded")

            const libArrMock = {
                libName: "test_library",
                uploadDate: "2024-12-01",
                nth: 0,
            }

            await libScraperInstance.moveToDownloadPage(newPage, libArrMock)
            expect(newPage.url()).toBe("https://www.data4library.kr/openDataV")
            newPage.waitForTimeout(1000)

            await libScraperInstance.downloadLibData(newPage, libArrMock)

            const isFileExist = fs.existsSync(
                path.join(process.env.DATA_PATH!, "library", libArrMock.libName, libArrMock.uploadDate + ".xlsx")
            )
            expect(isFileExist).toBe(true)
        },
        { timeout: 30_000 }
    )
})
