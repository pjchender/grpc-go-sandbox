package main

import (
	"context"
	"fmt"
	"log"
	"net"
	pb "sandbox/grpc-go-sandbox/routeguide"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// STEP 1-1：定義 routeGuideServer 的 struct
type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer
	savedFeatures []*pb.Feature
}

// STEP 1-2：根據 proto 中的 service 建立實作方式
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

func main() {
	// STEP 2-1：定義要監聽的 port 號
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 3000))
	if err != nil {
		log.Fatalf("failed to listed: %v", err)
	}

	// STEP 2-2：使用 gRPC 的 NewServer 方法來建立 gRPC Server 的實例
	grpcServer := grpc.NewServer()

	// STEP 2-3：在 gRPC Server 中註冊 service 的實作
	// 使用 proto 提供的 RegisterRouteGuideServer 方法，並將 routeGuideServer 作為參數傳入
	pb.RegisterRouteGuideServer(grpcServer, &routeGuideServer{})

	// STEP 2-4：啟動 grpcServer，並阻塞在這裡直到該程序被 kill 或 stop
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
