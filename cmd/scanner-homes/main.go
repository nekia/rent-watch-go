package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go"
	commondata "github.com/nekia/rent-watch-go/core/commondata"
	"github.com/playwright-community/playwright-go"
)

const (
	NATS_QUEUE_PREFIX = "room-"
	SITE_NAME         = "homes"
)

var (
	NATS_URL      = os.Getenv("NATS_SERVER_URL")
	USER_AGENT    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4595.0 Safari/537.36"
	WS_ENDPOINT   = os.Getenv("WS_ENDPOINT")
	WS_SESSION_ID = os.Getenv("WS_SESSION_ID")
)

func main() {

	if len(NATS_URL) == 0 || len(WS_ENDPOINT) == 0 || len(WS_SESSION_ID) == 0 {
		log.Fatalf("need to specify ws endpoint info")
	}

	// Connect to the NATS server
	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	c, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	chSend := make(chan *commondata.ScanResp)
	err = c.BindSendChan(commondata.NATS_SUBJECT_SCAN_RESP, chSend)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to a subject
	chRecv := make(chan *commondata.ScanReq)

	_, err = c.BindRecvQueueChan(commondata.NATS_SUBJECT_SCAN_REQ, NATS_QUEUE_PREFIX, chRecv)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Start waiting for scan req\n")

	// Wait for messages in a loop
	for msg := range chRecv {
		fmt.Printf("Received a msg from ch [%v]", msg)
		if msg.SiteName == SITE_NAME {
			fmt.Printf("Received message: %s\n", msg.Url)
			go scanRoomDetail(msg.Url)
		}
	}
}

func scanRoomDetail(url string) error {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	wsURL := path.Join(WS_ENDPOINT, WS_SESSION_ID)
	browser, err := pw.Chromium.Connect(wsURL)
	if err != nil {
		panic(err)
	}

	ignoreHttpsErrors := true
	ctx, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:         &USER_AGENT,
		IgnoreHttpsErrors: &ignoreHttpsErrors,
	})
	if err != nil {
		panic(err)
	}
	page, err := ctx.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(url); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	var resp commondata.ScanResp
	resp.Address, err = getAddress(&page)
	if err != nil {
		panic(err)
	}

	resp.BuiltYear, err = getBuiltYear(&page)
	if err != nil {
		panic(err)
	}

	resp.Price, err = getPrice(&page)
	if err != nil {
		panic(err)
	}

	resp.Size, err = getSize(&page)
	if err != nil {
		panic(err)
	}

	resp.FloorLevel = getFloorLevel(&page)

	resp.IsPetOK = getIsPetOK(&page)

	fmt.Printf("ScanResp: %v\n", resp)
	return nil
}

func getAddress(page *playwright.Page) (string, error) {
	pentry, err := (*page).QuerySelector("//th[text()='所在地']/following-sibling::td[1]//p")
	if err != nil {
		panic(err)
	}
	return pentry.TextContent()
}

func getBuiltYear(page *playwright.Page) (int, error) {
	pentries, err := (*page).QuerySelectorAll("//th[text()='築年月']/following-sibling::td[1]//p")
	if err != nil {
		panic(err)
	}
	fullBuiltDate, err := pentries[0].TextContent()
	if err != nil {
		panic(err)
	}
	split_fullBuiltDate := strings.Split(fullBuiltDate, "年")
	return strconv.Atoi(split_fullBuiltDate[0])
}

func getPrice(page *playwright.Page) (int, error) {
	pentry, err := (*page).QuerySelector("//th[text()='価格']/following-sibling::td[1]//p[1]/b")
	if err != nil {
		panic(err)
	}
	fullPriceStr, err := pentry.TextContent()
	if err != nil {
		panic(err)
	}
	split_fullPriceStr := strings.Split(fullPriceStr, "万")

	return strconv.Atoi(strings.Replace(split_fullPriceStr[0], ",", "", 1))
}

func getSize(page *playwright.Page) (float64, error) {
	pentry, err := (*page).QuerySelector("//th[text()='専有面積']/following-sibling::td[1]")
	if err != nil {
		panic(err)
	}
	fullSizeStr, err := pentry.TextContent()
	if err != nil {
		panic(err)
	}
	split_fullSizeStr := strings.Split(fullSizeStr, "㎡")
	return strconv.ParseFloat(split_fullSizeStr[0], 64)
}

func getFloorLevel(page *playwright.Page) commondata.FloorLevel {
	var fLevel commondata.FloorLevel

	pentry, err := (*page).QuerySelector("//th[text()='所在階 / 階数']/following-sibling::td[1]/span")
	if err != nil {
		panic(err)
	}
	fullFloorLevelStr, err := pentry.TextContent()
	if err != nil {
		panic(err)
	}

	split_fullFloorLevelStr := strings.Split(fullFloorLevelStr, "/")
	fLevel.FloorLevel, err = strconv.Atoi(strings.Replace(strings.TrimSpace(split_fullFloorLevelStr[0]), "階", "", 1))
	if err != nil {
		panic(err)
	}

	split_FloorTopLevel := strings.Split(strings.TrimSpace(split_fullFloorLevelStr[len(split_fullFloorLevelStr)-1]), "階")
	fLevel.FloorTopLevel, err = strconv.Atoi(split_FloorTopLevel[0])
	if err != nil {
		panic(err)
	}

	return fLevel
}

func getIsPetOK(page *playwright.Page) bool {
	pentry, err := (*page).QuerySelector("//th[text()='その他条件']/following-sibling::td[1]")
	if err != nil {
		panic(err)
	}
	fullCondition, err := pentry.TextContent()
	if err != nil {
		panic(err)
	}

	if strings.Contains(fullCondition, "ペット可") {
		return true
	} else if strings.Contains(fullCondition, "ペット相談") {
		return true
	} else {
		return false
	}

}
