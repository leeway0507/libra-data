import { kyoboScraper } from "./book-scraper"

async function main() {
    const isbnArr = [
        // "9791156000846",
        // "9791193217078",
        // "9791197932564",
        // "9788980783144",
        // "9791192987354",
        // "9788980783144",
        "9791158394622",
        // "9788957273647",
        // "9791165711856",
        // "9791193926260",
        // "9791169212168",
    ]
    const scraper = new kyoboScraper()
    const x = await Promise.all((isbnArr.map(async (isbn) => {
        return await scraper.exec(isbn)
    })))

    console.log(x)

}

main().catch((err) => {
    console.error("Error occurred:", err)
}).finally(() => {
    process.exit(0) // 프로세스 강제 종료
})
