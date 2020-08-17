package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	pb "sandbox/grpc-go-sandbox/routeguide"
	"time"
)

var serverAddr = "localhost:3000"

// printFeature gets the feature for the given point
func printFeature(client pb.RouteGuideClient, point *pb.Point) {
	log.Printf("Getting feature for point (%d, %d)", point.Latitude, point.Longitude)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feature, err := client.GetFeature(ctx, point)
	if err != nil {
		log.Fatalf("%v.GetFeature(_) = _, %v: ", client, err)
	}
	log.Println(feature)
}

func main() {
	// STEP 1：creating the client stub
	// STEP 1-1：與 gRPC server 建立 channel
	// 如果沒有使用安全連線的話，在 options 的地方要加上 grpc.WIthInsecure()
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	// STEP 1-2：使用 proto 所提供的 NewRouteGuideClient 方法並帶入參數 conn 以來建立 client
	client := pb.NewRouteGuideClient(conn)

	// STEP 2：呼叫 Service Methods
	printFeature(client, &pb.Point{Latitude: 409146138, Longitude: -746188906})
}
