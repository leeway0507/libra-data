import { kyoboScraper } from "./book-scraper";

describe(('Consortium List SubScraper '), () => {
    const scraper = new kyoboScraper();
    const testIsbn = "9791163034735" // 점프 투 파이썬
    scraper.isbn = testIsbn

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

    test('should saveBookHtml ', () => {
        scraper.saveHtml();
        expect(scraper.page.url).toEqual("")
    });

    test('should saveImage ', () => {
        expect(scraper.saveImage(Buffer.from("test"))).toEqual("ok")
    });

    test('should searchBook ', async () => {
        await scraper.initBrowser()
        const testPath = "file:///Users/yangwoolee/repo/libra-data/scraper/temp/test/kyobo/search.html"
        const url = await scraper.searchBook(testPath);

        expect(url).toEqual("https://product.kyobobook.co.kr/detail/S000202532365")
    });

    test('should getRecommendation ', () => {
        expect(scraper.getRecommendation()).toEqual("")
    });

    test('should getToc ', () => {
        expect(scraper.getToc()).toEqual("")
    });
})