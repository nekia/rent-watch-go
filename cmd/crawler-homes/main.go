package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/playwright-community/playwright-go"
)

const (
	NATS_SUBJECT_CRAWL_REQ  = "crawl-request"
	NATS_SUBJECT_CRAWL_RESP = "crawl-response"
	NATS_QUEUE_PREFIX       = "room-"
	SITE_NAME               = "homes"
)

var (
	NATS_URL      = os.Getenv("NATS_SERVER_URL")
	URL           = "https://www.homes.co.jp/mansion/chuko/tokyo/list/?cond%5Bcity%5D%5B13104%5D=13104&cond%5Bcity%5D%5B13113%5D=13113&cond%5Bcity%5D%5B13110%5D=13110&cond%5Bcity%5D%5B13114%5D=13114&cond%5Bmoneyroom%5D=0&cond%5Bmoneyroomh%5D=10000&cond%5Bhousearea%5D=60&cond%5Bhouseareah%5D=0&cond%5Bwalkminutesh%5D=0&cond%5Bhouseageh%5D=0&cond%5Bmcf%5D%5B340102%5D=340102&cond%5Bmcf%5D%5B113201%5D=113201&bukken_attr%5Bcategory%5D=mansion&bukken_attr%5Bbtype%5D=chuko&bukken_attr%5Bpref%5D=13"
	USER_AGENT    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4595.0 Safari/537.36"
	WS_ENDPOINT   = os.Getenv("WS_ENDPOINT")
	WS_SESSION_ID = os.Getenv("WS_SESSION_ID")
)

type CrawlReq struct {
	SiteName string `json:"siteName,omitempty"`
}

type CrawlResp struct {
	Url string `json:"url,omitempty"`
}

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

	// Subscribe to a subject
	chRecv := make(chan *CrawlReq)
	_, err = c.BindRecvChan(NATS_SUBJECT_CRAWL_REQ, chRecv)
	if err != nil {
		log.Fatal(err)
	}

	chSend := make(chan *CrawlResp)
	err = c.BindSendChan(NATS_SUBJECT_CRAWL_RESP, chSend)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Start waiting for crawl request... [%s]\n", SITE_NAME)

	for msg := range chRecv {
		if msg.SiteName == SITE_NAME {
			startCrawl(chSend)
		}
	}

}

func startCrawl(ch chan *CrawlResp) error {
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
	if _, err = page.Goto(URL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	for {
		filename := fmt.Sprintf("./screenshot-%d.png", time.Now().Unix())
		_, err = page.Screenshot(playwright.PageScreenshotOptions{Path: &filename})
		if err != nil {
			panic(err)
		}
		entries, err := page.QuerySelectorAll("//div[contains(@class, 'cMansion') and contains(@class, 'mod-mergeBuilding--sale')]//td[@class='detail']")
		if err != nil {
			log.Fatalf("could not get entries: %v", err)
		}

		for i, entry := range entries {
			detailLinkElement, err := entry.QuerySelector("a")
			if err != nil {
				log.Fatalf("could not get title element: %v", err)
			}
			detailLink, err := detailLinkElement.GetAttribute("href")
			if err != nil {
				log.Fatalf("could not get text content: %v", err)
			}
			fmt.Printf("%d: %s\n", i+1, detailLink)

			ch <- &CrawlResp{Url: detailLink}
		}

		if err = pagination(&browser, &page); err != nil {
			break
		}
	}
	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}

	return nil
}

func pagination(browser *playwright.Browser, page *playwright.Page) error {
	entry, err := (*page).QuerySelector("//div[contains(@class, 'mod-listPaging')]//li[@class='nextPage']")
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	} else if entry == nil {
		fmt.Println("End of pages")
		return errors.New("end of pages")
	}

	selectedPageElm, err := (*page).QuerySelector("//div[contains(@class, 'mod-listPaging')]//li[@class='selected']/span")
	if err != nil {
		log.Fatalf("could not get entry: %v", err)
	}

	currentPageStr, err := selectedPageElm.TextContent()
	if err != nil {
		log.Fatalf("could not get text: %v", err)
	}

	i, err := strconv.Atoi(currentPageStr)
	if err != nil {
		log.Fatalf("could not convert to integer: %v", err)
	}

	nextPageURL := fmt.Sprintf("%s&page=%d", URL, i+1)
	fmt.Printf("%s\n", nextPageURL)

	(*page).Close()

	ignoreHttpsErrors := true
	ctx, err := (*browser).NewContext(playwright.BrowserNewContextOptions{
		UserAgent:         &USER_AGENT,
		IgnoreHttpsErrors: &ignoreHttpsErrors,
	})
	if err != nil {
		panic(err)
	}

	*page, err = ctx.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	if _, err = (*page).Goto(nextPageURL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	return nil
}
