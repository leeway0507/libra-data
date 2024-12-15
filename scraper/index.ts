import {
    scrapIsbns,
    saveScrapResult,
    loadTargets,
    updateTargetResult,
} from "./book-scraper"

async function main() {
    const targetIsbns = await loadTargets().then((a) => a.slice(0, 40))
    const scrapResult = await scrapIsbns(targetIsbns, 4)
    const resultIsbns = await saveScrapResult(scrapResult)

    console.log("scrap Length :", resultIsbns.length)
    await updateTargetResult(targetIsbns, resultIsbns)
}

main()
    .catch((err) => {
        console.error("Error occurred:", err)
    })
    .finally(() => {
        process.exit(0) // 프로세스 강제 종료
    })
