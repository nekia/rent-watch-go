package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
	crawler "github.com/nekia/rent-watch-go/cmd/crawler-homes"
	scanner "github.com/nekia/rent-watch-go/cmd/scanner-homes"
)

var (
	ENDPOINT_URL = "https://notify-api.line.me/api/notify"
	LINE_TOKEN   = os.Getenv("LINE_NOTIFY_TOKEN")
	NATS_URL     = os.Getenv("NATS_SERVER_URL")
)

const (
	NATS_SUBJECT_CRAWL_REQ    = "crawl-request"
	NATS_SUBJECT_CRAWL_RESP   = "crawl-response"
	NATS_SUBJECT_SCANNER_REQ  = "scanner-request"
	NATS_SUBJECT_SCANNER_RESP = "scanner-response"
	NATS_QUEUE_PREFIX         = "room-"
	SITE_NAME                 = "homes"
)

func startReceiveScanResults(ch chan *scanner.ScanResp) {
	for msg := range ch {
		// Send a scan request against each room detail site (NATS)
		fmt.Printf("Receiv Scan Resp to [%s]\n", msg.Location)
	}
}

func main() {

	if len(NATS_URL) == 0 {
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
	chCrawlerSend := make(chan *crawler.CrawlReq)
	err = c.BindSendChan(NATS_SUBJECT_CRAWL_REQ, chCrawlerSend)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to a subject
	chCrawlerRecv := make(chan *crawler.CrawlResp)
	_, err = c.BindRecvChan(NATS_SUBJECT_CRAWL_RESP, chCrawlerRecv)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to a subject
	chScannerSend := make(chan *scanner.ScanReq)
	err = c.BindSendChan(NATS_SUBJECT_SCANNER_REQ, chScannerSend)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to a subject
	chScannerRecv := make(chan *scanner.ScanResp)
	_, err = c.BindRecvChan(NATS_SUBJECT_SCANNER_RESP, chScannerRecv)
	if err != nil {
		log.Fatal(err)
	}

	// Send a request to crawler (NATS)
	chCrawlerSend <- &crawler.CrawlReq{SiteName: "homes"}

	go startReceiveScanResults(chScannerRecv)

	// Receive a response from crawler
	for msg := range chCrawlerRecv {
		// Send a scan request against each room detail site (NATS)
		chScannerSend <- &scanner.ScanReq{SiteName: "homes", Url: msg.Url}
		fmt.Printf("Send Scan Req to [%s]\n", msg.Url)
	}
	// Send multiple requests to the scanner simultanously

	// Call grpc API to check if a room condition satisfies criterias configured or not

	// Only for the room passed the previous check, Send a notification request (NATS JS)

}
