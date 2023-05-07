package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	pb "github.com/nekia/rent-watch-go/protobuf/checker"
)

func main() {

	// Connect to NATS
	nc, _ := nats.Connect(nats.DefaultURL)

	// Create JetStream Context
	js, _ := nc.JetStream()

	path := filepath.Join("test-roomdetail.json")
	jsonText, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var detail pb.RoomDetail
	json.Unmarshal([]byte(jsonText), &detail)
	fmt.Printf("%v\n", detail)

	data, err := proto.Marshal(&detail)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// // Simple Stream Publisher
	js.Publish("roomdetails", data)

}
