import { scrapIsbns, saveScrapResult, loadTargets, updateTargetResult, initBrowser } from "./book-scraper"
import { LibScraper } from "./library-scraper"



async function main() {
    await scrapLibraryData()
}
main()
    .catch((err) => {
        console.error("Error occurred:", err)
    })
    .finally(() => {
        process.exit(0) // 프로세스 강제 종료
    })


async function scrapBookDataFromKyobo() {
    for (let index = 0; index < 1; index++) {
        const targetIsbns = await loadTargets().then((a) => a.slice(0, 8))
        const scrapResult = await scrapIsbns(targetIsbns, 8, true)
        const resultIsbns = await saveScrapResult(scrapResult)

        console.log("scrap Length :", resultIsbns.length)
        await updateTargetResult(targetIsbns, resultIsbns)
    }
}
async function scrapLibraryData() {
    const ctx = await initBrowser()
    for (let pageNum = 1; pageNum <= 10; pageNum++) {
        const libScraperInstance = new LibScraper(ctx)

        const page = await ctx.newPage()
        await libScraperInstance.getDataByPagination(page, pageNum)
        await page.close()
    }
}