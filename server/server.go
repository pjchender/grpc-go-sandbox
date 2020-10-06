package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	pb "sandbox/grpc-go-sandbox/routeguide"
	"time"

	"sandbox/grpc-go-sandbox/testdata"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
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

// ListFeatures 會以 server-side stream 的方式將所有 features 回傳給 client
func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	// 取出 savedFeatures 的資料並以 stream.Send() 的方式回傳
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

func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()

	for {
		// 在 server 端接收 client 傳來的 stream
		point, err := stream.Recv()
		log.Println(point)

		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.RouteSummary{
				PointCount:   pointCount,
				FeatureCount: featureCount,
				Distance:     distance,
				ElapsedTime:  int32(endTime.Sub(startTime).Seconds()),
			})
		}

		if err != nil {
			return err
		}

		pointCount++
		for _, feature := range s.savedFeatures {
			if proto.Equal(feature.Location, point) {
				featureCount++
			}
		}
		if lastPoint != nil {
			distance += calcDistance(lastPoint, point)
		}
	}
}

func main() {
	addr := fmt.Sprintf("localhost:%d", 3000)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	grpcServer := grpc.NewServer()
	s := &routeGuideServer{}

	// 把 route_guide_db.json 的資料載入
	dataPath := testdata.Path("route_guide_db.json")
	s.loadFeatures(dataPath)

	// 使用 proto 提供的 RegisterRouteGuideServer 方法，並將 routeGuideServer 作為參數傳入
	// 把 s 傳入 grpcServer
	pb.RegisterRouteGuideServer(grpcServer, s)

	fmt.Printf("grpcServer started at %v", addr)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// 判斷 Point 是否在 Rectangle 的區域內
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

func toRadians(num float64) float64 {
	return num * math.Pi / float64(180)
}

// calcDistance calculates the distance between two points using the "haversine" formula.
// The formula is based on http://mathforum.org/library/drmath/view/51879.html.
func calcDistance(p1 *pb.Point, p2 *pb.Point) int32 {
	const CordFactor float64 = 1e7
	const R = float64(6371000) // earth radius in metres
	lat1 := toRadians(float64(p1.Latitude) / CordFactor)
	lat2 := toRadians(float64(p2.Latitude) / CordFactor)
	lng1 := toRadians(float64(p1.Longitude) / CordFactor)
	lng2 := toRadians(float64(p2.Longitude) / CordFactor)
	dlat := lat2 - lat1
	dlng := lng2 - lng1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	return int32(distance)
}
