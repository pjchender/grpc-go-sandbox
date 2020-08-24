package main

import (
	"context"
	"google.golang.org/grpc"
	"io"
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

// STEP 1：撰寫 PrintFeatures 取得 server-side streaming gRPC 的資料
func printFeatures(client pb.RouteGuideClient, rect *pb.Rectangle) {
	log.Printf("Looking for feature with %v", rect)

	// STEP 2：透過 context.WithTimeout 建立 timeout 機制，並取得 ctx
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// STEP 3：將 ctx 傳入 pb 提供的 ListFeatures 方法，可以得到 stream
	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}

	// STEP 4：透過 for loop 搭配 stream.Recv() 方法可以取得每次串流的資料
	for {
		feature, err := stream.Recv()

		// STEP 4-1：io.EOF 表示資料讀完了
		if err == io.EOF {
			break
		}

		// STEP 4-2: error handling
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}

		// STEP 4-3: print feature
		log.Println(feature)
	}
}

func main() {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRouteGuideClient(conn)

	// A simple gRPC
	//printFeature(client, &pb.Point{Latitude: 409146138, Longitude: -746188906})

	// Server-side streaming RPC
	// STEP 5: use print features function
	printFeatures(client, &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	})
}
