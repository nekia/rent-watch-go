package main

import (
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

const (
	NATS_SUBJECT_SCAN_REQ  = "scan-request"
	NATS_SUBJECT_SCAN_RESP = "scan-response"
	NATS_QUEUE_PREFIX      = "room-"
	SITE_NAME              = "homes"
)

var (
	NATS_URL      = os.Getenv("NATS_SERVER_URL")
	USER_AGENT    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4595.0 Safari/537.36"
	WS_ENDPOINT   = os.Getenv("WS_ENDPOINT")
	WS_SESSION_ID = os.Getenv("WS_SESSION_ID")
)

type floorLevel struct {
	FloorLevel    int `json:"floorLevel,omitempty"`
	FloorTopLevel int `json:"floorTopLevel,omitempty"`
}
type scanResp struct {
	Address    string     `json:"address,omitempty"`
	Price      float32    `json:"price,omitempty"`
	Size       float32    `json:"size,omitempty"`
	FloorLevel floorLevel `json:"floorLevel,omitempty"`
	Location   string     `json:"location,omitempty"`
	BuiltYear  int        `json:"builtYear,omitempty"`
	IsPetOK    bool       `json:"isPetOK,omitempty"`
}

type scanReq struct {
	SiteName string `json:"siteName,omitempty"`
	Url      string `json:"url,omitempty"`
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

	chSend := make(chan *scanResp)
	err = c.BindSendChan(NATS_SUBJECT_SCAN_RESP, chSend)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe to a subject
	chRecv := make(chan *scanReq)

	_, err = c.BindRecvQueueChan(NATS_SUBJECT_SCAN_REQ, NATS_QUEUE_PREFIX, chRecv)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for messages in a loop
	for msg := range chRecv {
		if msg.SiteName == SITE_NAME {
			fmt.Printf("Received message: %s\n", msg.Url)
			scanRoomDetail(msg.Url)
		}
	}
}

func scanRoomDetail(url string) error {
	return nil
}
