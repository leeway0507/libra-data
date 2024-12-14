import { scrapIsbns } from "./book-scraper"

async function main() {
    const isbnArr = [
        "9791156000846",
        "9791193217078",
        "9791197932564",
        "9788980783144",
        "9791192987354",
        "9788980783144",
        "9791158394622",
    ]
    const x = await scrapIsbns(isbnArr, 3)
    console.log(x)
}

main().catch((err) => {
    console.error("Error occurred:", err)
}).finally(() => {
    process.exit(0) // 프로세스 강제 종료
})
