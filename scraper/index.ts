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
async function scrapLibData() {
    const ctx = await initBrowser()
    const page = await ctx.newPage()
    const libScraperInstance = new LibScraper(ctx)
    await libScraperInstance.getDataByPagination(page, 1)
}


async function main() {
    await scrapLibData()
}
main()
    .catch((err) => {
        console.error("Error occurred:", err)
    })
    .finally(() => {
        process.exit(0) // 프로세스 강제 종료
    })

