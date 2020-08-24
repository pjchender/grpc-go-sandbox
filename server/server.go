package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	pb "sandbox/grpc-go-sandbox/routeguide"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"sandbox/grpc-go-sandbox/testdata"
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
	savedFeatures []*pb.Feature
}

// 在 proto 中有定義這個 service 會接收 point 最為參數，並且會回傳 Feature
func (s *routeGuideServer) GetFeature(ctx context.Context, point *pb.Point) (
	*pb.Feature, error,
) {
	for _, feature := range s.savedFeatures {
		if proto.Equal(feature.Location, point) {
			return feature, nil
		}
	}

	// No feature was found, return an unnamed feature
	log.Println("No feature was found, return an unnamed feature.")
	return &pb.Feature{Location: point}, nil
}

// STEP 1：ListFeatures 會以 server-side stream 的方式將所有 features 回傳給 client
func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	// STEP 2：取出 savedFeatures 的資料並以 stream.Send() 的方式回傳
	// Client 端需要從 err 判斷，如果還有資料未傳完，則 err 會是 nil；如果傳完了會是 io.EOF；否則會得到 err
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		} else {
			fmt.Println("out of range")
		}
	}
	return nil
}

// 定義 loadFeatures 會把 JSON 檔案中所列的 features 載入
func (s *routeGuideServer) loadFeatures(filePath string) {
	var data []byte

	if filePath != "" {
		var err error
		data, err = ioutil.ReadFile(filePath)

		if err != nil {
			log.Fatalf("Failed to load default features: %v", err)
		}
	} else {
		//data = exampleData
		log.Fatalf("filePath is not exists")
	}

	// 將 bytes 轉成 struct
	if err := json.Unmarshal(data, &s.savedFeatures); err != nil {
		log.Fatalf("Failed to load default features: %v", err)
	}
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 3000))
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	grpcServer := grpc.NewServer()
	s := &routeGuideServer{}

	// STEP 2：把 route_guide_db.json 的資料載入
	dataPath := testdata.Path("route_guide_db.json")
	s.loadFeatures(dataPath)

	// 使用 proto 提供的 RegisterRouteGuideServer 方法，並將 routeGuideServer 作為參數傳入
	// STEP 3：把 s 傳入 grpcServer
	pb.RegisterRouteGuideServer(grpcServer, s)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// STEP 2：判斷 Point 是否在 Rectangle 的區域內
func inRange(point *pb.Point, rect *pb.Rectangle) bool {
	left := math.Min(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	right := math.Max(float64(rect.Lo.Longitude), float64(rect.Hi.Longitude))
	top := math.Max(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))
	bottom := math.Min(float64(rect.Lo.Latitude), float64(rect.Hi.Latitude))

	if float64(point.Longitude) >= left &&
		float64(point.Longitude) <= right &&
		float64(point.Latitude) >= bottom &&
		float64(point.Latitude) <= top {
		return true
	}

	return false
}
