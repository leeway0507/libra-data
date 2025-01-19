import { BrowserContext, Page } from "playwright"
import path from "path"
import fsSync from "fs"
import fsAsync from "fs/promises"

if (!process.env.DATA_PATH) {
    process.env.DATA_PATH = "/Users/yangwoolee/repo/libra-data/data"
}

type LibScrap = {
    libName: string
    uploadDate: string
    nth: number
}

// 도서관 정보나루의 데이터 제공 페이지, 도서관 별 장서 대출 xlsx 파일 수집
export class LibScraper {
    ctx!: BrowserContext //playwright

    constructor(page: BrowserContext) {
        this.ctx = page
    }

    // 현재 페이지의 도서관 데이터 방문 및 파일 다운로드
    async getDataByPagination(page: Page, targetPage: number) {
        await this.openPage(page)
        await this.setFilterOption(page)
        await this.moveToTargetPagination(page, targetPage)
        const libList: LibScrap[] = await this.getLibList(page)
        const libFilteredList: LibScrap[] = this.filterAlreadyScrapLib(libList)

        if (libFilteredList.length > 0) {
            for (const libScrap of libFilteredList) {
                await this.moveToDownloadPage(page, libScrap)
                await this.downloadLibData(page, libScrap)
                await page.goBack()
                await page.waitForLoadState("networkidle")
            }
        }
    }

    async openPage(page: Page) {
        await page.goto("https://www.data4library.kr/openDataL")
        await page.waitForLoadState("domcontentloaded")
    }

    async setFilterOption(page: Page) {
        await this.setLocationToSeoul(page)
        await this.setLibTypeToPublic(page)
    }

    async setLocationToSeoul(page: Page): Promise<boolean> {
        const x = await page.selectOption("#p_region", "서울")
        await page.waitForLoadState("networkidle")
        return x.length > 0 && x[0] === "11"
    }

    async setLibTypeToPublic(page: Page): Promise<boolean> {
        const x = await page.selectOption("#libType", "공공")
        await page.waitForLoadState("networkidle")
        return x.length > 0
    }

    async moveToTargetPagination(page: Page, targetPage: number) {
        if (targetPage > 5) {
            const nextPaginationXPath = "//a[contains(@class, 'next_page')]"
            const loc = page.locator(nextPaginationXPath)
            await loc.first().click()
            await page.waitForLoadState("networkidle")
        }
        const pageNationXPath = "//a[@class='page']"
        const loc = page.locator(pageNationXPath)
        const idx = (targetPage % 5) - 1
        await loc.nth(idx === -1 ? 4 : idx).click()
        await page.waitForLoadState("networkidle")
    }

    // 현재 페이지 내 존재하는 도서관 목록 확보
    async getLibList(page: Page): Promise<LibScrap[]> {
        const tableXpath = "//tbody/tr"
        const libNameXpath = "//td[@class='link_td']/a"
        const uploadDateXpath = "//td[@class='br_none']"
        const libList = await page.locator(tableXpath).all()
        return Promise.all(
            libList.map(async (v, idx) => {
                const lib1 = v.locator(libNameXpath)
                const lib2 = v.locator(uploadDateXpath)
                const libName = (await lib1.count()) > 0 && (await lib1.innerText())
                const uploadDate = (await lib2.count()) > 0 && (await lib2.innerText())
                return {
                    libName: libName ? libName.replace("장서/대출 데이터", "").trim() : "",
                    uploadDate: uploadDate || "",
                    nth: idx,
                }
            })
        )
    }

    // 도서관 목록에서 기존 수집 완료된 도서관 제거
    filterAlreadyScrapLib(libList: LibScrap[]): LibScrap[] {
        const isCandidate = (l: LibScrap) => {
            const folderPath = path.join(process.env.DATA_PATH!, "library", l.libName)
            const filePath = path.join(folderPath, l.uploadDate + ".xlsx")
            return !fsSync.existsSync(filePath)
        }

        return libList.filter(isCandidate)
    }

    async moveToDownloadPage(page: Page, libData: LibScrap) {
        const libNameXpath = "//td[@class='link_td']/a"
        await page.locator(libNameXpath).nth(libData.nth).click()
        await page.waitForLoadState("networkidle")
    }

    async downloadLibData(page: Page, libData: LibScrap) {
        const dateXpath = "//section/div[2]/div[3]/table/tbody/tr[1]/td[3]"
        const loc = page.locator(dateXpath)
        if ((await loc.innerText()) !== libData.uploadDate) {
            console.log("not match", libData)
            return
        }

        const downloadXpath = "//section/div[2]/div[3]/table/tbody/tr[1]/td[4]/a[2]"
        const downloadLoc = page.locator(downloadXpath)

        await fsAsync.mkdir(process.env.DATA_PATH!, { recursive: true })

        const downloadPromise = page.waitForEvent("download")
        await downloadLoc.click()
        const download = await downloadPromise
        const folderPath = path.join(process.env.DATA_PATH!, "library", libData.libName)
        await download.saveAs(path.join(folderPath, libData.uploadDate + ".xlsx"))
    }
}
