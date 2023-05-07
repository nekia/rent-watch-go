package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	pb "github.com/nekia/rent-watch-go/protobuf/checker"
)

var (
	ENDPOINT_URL = "https://notify-api.line.me/api/notify"
	LINE_TOKEN   = os.Getenv("LINE_NOTIFY_TOKEN")
)

func main() {

	// Connect to NATS
	nc, _ := nats.Connect(nats.DefaultURL)

	// Create JetStream Context
	js, _ := nc.JetStream(nats.PublishAsyncMaxPending(256))

	// Create a stream
	stream, err := js.AddStream(&nats.StreamConfig{
		Name:     "mystream",
		Subjects: []string{"roomdetails"},
		MaxBytes: -1,
		Storage:  nats.MemoryStorage,
		Discard:  nats.DiscardOld,
		Replicas: 1,
	})
	if err != nil {
		panic((err))
	}

	// Create a Consumer
	_, err = js.AddConsumer(stream.Config.Name, &nats.ConsumerConfig{
		Durable:       "myconsumer",
		FilterSubject: "roomdetails",
		AckPolicy:     nats.AckExplicitPolicy,
	})
	if err != nil {
		panic((err))
	}

	sub, err := js.PullSubscribe("roomdetails", "myconsumer", nats.BindStream("mystream"))
	if err != nil {
		panic(err)
	}

	i := 0
	for {
		msgs, _ := sub.Fetch(10, nats.MaxWait(60*time.Second))
		for _, msg := range msgs {
			var detail pb.RoomDetail
			proto.Unmarshal(msg.Data, &detail)
			fmt.Printf("%d:%s\n", i, detail.GetAddress())
			msg.Ack()
			i++
			notifyLINE(&detail)
		}
	}
}

func notifyLINE(detail *pb.RoomDetail) {
	// Request payload
	var message string
	if detail.GetIsPetOK() {
		message += "ペットOK\n"
	}
	message += fmt.Sprintf("%.1f万円 %.2f平米 %d/%d\n",
		detail.GetPrice(), detail.GetSize(), detail.FloorLevel.GetFloorLevel(), detail.GetFloorLevel().FloorTopLevel)
	message += fmt.Sprintf("%s\n%s", detail.GetAddress(), detail.GetUrl())

	payload := url.Values{}
	payload.Set("message", message)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", ENDPOINT_URL, strings.NewReader(payload.Encode()))
	if err != nil {
		fmt.Printf("Failed to create a request\n")
		panic(err)
		// log.Fatal(err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+LINE_TOKEN)

	// Send the HTTP request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to do client\n")
		panic(err)
		// log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status code: %d", resp.StatusCode)
	}
}
