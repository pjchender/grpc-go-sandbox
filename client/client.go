package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	pb "sandbox/grpc-go-sandbox/routeguide"
)

var serverAddr = "localhost:3000"

func main() {
	// STEP 1：creating the client stub
	// STEP 1-1：與 gRPC server 建立 channel
	// 如果沒有使用安全連線的話，在 options 的地方要加上 grpc.WIthInsecure()
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("faild to dial: %v", err)
	}
	defer conn.Close()

	// STEP 1-2：使用 proto 所提供的 NewRouteGuideClient 方法並帶入參數 conn 以來建立 client
	client := pb.NewRouteGuideClient(conn)

	// STEP 2：呼叫 Service Methods
	// 透過 context.Context 物件，讓我們在需要時可以改變 RPC 的行為，像是立即執行 time-out/cancel 一個 RPC
	feature, err := client.GetFeature(context.Background(), &pb.Point{Latitude: 409146138, Longitude: -746188906})
	if err != nil {
		log.Fatalf("faild to getFeature")
	}

	log.Println(feature)
}
