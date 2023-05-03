package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pb "github.com/nekia/rent-watch-go/protobuf/checker"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	rdb       *redis.Client
	condition *checkCondition
	pb.UnimplementedCheckerServer
}

// CheckByUrl implements checker.checkerServer
func (s *server) CheckByUrl(ctx context.Context, in *pb.CheckByUrlRequest) (*pb.CheckByUrlResponse, error) {
	url := in.GetUrl()
	log.Printf("Received: %v", url)
	_, err := s.rdb.Get(ctx, url).Result()
	if err == redis.Nil {
		return &pb.CheckByUrlResponse{Result: pb.CheckStatus_NOT_INSPECTED}, nil
	} else if err == nil {
		return &pb.CheckByUrlResponse{Result: pb.CheckStatus_ALREADY_INSPECTED}, nil
	} else {
		panic(err)
	}
}

// CheckByRoomDetail implements checker.checkerServer
func (s *server) CheckByRoomDetail(ctx context.Context, in *pb.CheckByRoomDetailRequest) (*pb.CheckByRoomDetailResponse, error) {
	detail := in.GetDetail()
	index := composeIndex(detail)
	log.Printf("Received: %v(%s)", detail, index)
	fmt.Println(s.condition)

	_, err := s.rdb.Get(ctx, index).Result()
	if err == redis.Nil {
		log.Printf("Continue...")
	} else if err == nil {
		return &pb.CheckByRoomDetailResponse{Result: pb.CheckStatus_ALREADY_INSPECTED}, nil
	} else {
		panic(err)
	}

	evalResult := evaluateRoomDetail(s.condition, detail)
	if evalResult {
		return &pb.CheckByRoomDetailResponse{Result: pb.CheckStatus_SATISFIED}, nil
	} else {
		return &pb.CheckByRoomDetailResponse{Result: pb.CheckStatus_NOT_SATISFIED}, nil
	}
}

// UpdateCheckStatus implements checker.checkerServer
func (s *server) UpdateCheckStatus(ctx context.Context, in *pb.UpdateCheckStatusRequest) (*pb.UpdateCheckStatusResponse, error) {
	detail := in.GetDetail()
	index := composeIndex(detail)
	status := in.GetStatus()
	log.Printf("Received: %v(%s)%d", detail, index, status)
	fmt.Println(s.condition)

	_, err := s.rdb.Set(ctx, index, int32(status.Enum().Number()), 0).Result()
	if err != nil {
		panic(err)
	}

	_, err = s.rdb.Set(ctx, detail.GetUrl(), int32(status.Enum().Number()), 0).Result()
	if err != nil {
		panic(err)
	}

	return &pb.UpdateCheckStatusResponse{Result: status}, nil
}

func evaluateRoomDetail(condition *checkCondition, detail *pb.RoomDetail) bool {
	if detail.GetPrice() > condition.MaxPrice {
		log.Printf("Too expensive - %.2f", detail.GetPrice())
		return false
	} else if detail.GetPrice() < condition.MinPrice {
		log.Printf("Too cheep - %.2f", detail.GetPrice())
		return false
	} else if int32(time.Now().Year())-detail.GetBuiltYear() > condition.MaxBuiltAge {
		log.Printf("Too old - %d", detail.GetBuiltYear())
		return false
	} else if detail.GetSize() < condition.MinSize {
		log.Printf("Too small - %.2f", detail.GetSize())
		return false
	} else if detail.GetFloorLevel().GetFloorLevel() < condition.MinFloorLevel {
		log.Printf("Too low floor - %d", detail.FloorLevel.GetFloorLevel())
		return false
	} else if detail.GetFloorLevel().GetFloorLevel() == detail.GetFloorLevel().GetFloorTopLevel() {
		log.Printf("At the top of floor - %d / %d", detail.GetFloorLevel().GetFloorLevel(), detail.GetFloorLevel().GetFloorTopLevel())
		return false
	}
	return true
}

func composeIndex(detail *pb.RoomDetail) string {
	indexStr := strings.Join([]string{
		detail.GetAddress(),
		strconv.FormatInt(int64(detail.GetBuiltYear()), 10),
		strconv.FormatFloat(detail.GetPrice(), 'f', 1, 32),
		strconv.FormatFloat(detail.GetSize(), 'f', 2, 32),
		strconv.FormatInt(int64(detail.GetFloorLevel().GetFloorLevel()), 10),
		strconv.FormatInt(int64(detail.GetFloorLevel().GetFloorTopLevel()), 10)}, "-")
	return indexStr
}

type checkCondition struct {
	MaxPrice      float64 `json:"max_room_price"`
	MinPrice      float64 `json:"min_room_price"`
	MinSize       float64 `json:"min_room_size"`
	MinFloorLevel int32   `json:"min_floor_level"`
	MaxBuiltAge   int32   `json:"max_building_age"`
	MinRank       int32   `json:"min_rank"`
}

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	path := filepath.Join("config.json")
	jsonText, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	condition := checkCondition{}
	json.Unmarshal([]byte(jsonText), &condition)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCheckerServer(s, &server{rdb: rdb, condition: &condition})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
