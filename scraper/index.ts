import { scrapIsbns, saveScrapResult, loadTargets, updateTargetResult, randomNumber, initBrowser } from "./book-scraper"
import { LibScraper } from "./library-scraper"

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
    for (let i = 6; i <= 10; i++) {
        const page = await ctx.newPage()
        const libScraperInstance = new LibScraper(ctx)
        await libScraperInstance.getDataByPagination(page, i)
        await page.close()
    }
}

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
