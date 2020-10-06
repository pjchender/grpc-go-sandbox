package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	pb "sandbox/grpc-go-sandbox/routeguide"
	"time"

	"google.golang.org/grpc"
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

// 撰寫 PrintFeatures 取得 server-side streaming gRPC 的資料
func printFeatures(client pb.RouteGuideClient, rect *pb.Rectangle) {
	log.Printf("Looking for feature with %v", rect)

	// 透過 context.WithTimeout 建立 timeout 機制，並取得 ctx
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 將 ctx 傳入 pb 提供的 ListFeatures 方法，可以得到 stream
	stream, err := client.ListFeatures(ctx, rect)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}

	// 透過 for loop 搭配 stream.Recv() 方法可以取得每次串流的資料
	for {
		feature, err := stream.Recv()

		// io.EOF 表示資料讀完了
		if err == io.EOF {
			break
		}

		// error handling
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}

		// print feature
		log.Println(feature)
	}
}

// client-to-server stream
// runRecordRoute 會送一系列的 points 到 server，並從 server 取得 RouteSummary 的回應
func runRecordRoute(client pb.RouteGuideClient) {
	// 建立隨機 points
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	pointCount := int(r.Int31n(100)) + 2
	var points []*pb.Point
	for i := 0; i < pointCount; i++ {
		points = append(points, randomPoint(r))
	}
	log.Printf("Traversing %d points.", len(points))

	// 建立 timeout 機制
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 透過 RouteGuideClient 的 RecordRoute 方法可以取得 stream
	stream, err := client.RecordRoute(ctx)

	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
	}

	for _, point := range points {
		// 透過 stream.Send 向 server 發送 stream
		if err := stream.Send(point); err != nil {
			log.Fatalf("%v.Send(%v) = %v", stream, point, err)
		}
	}

	// 告知 server 傳送完畢，並準備接收 server 的回應
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	log.Printf("Route summary: %v", reply)
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

	// server-to-client streaming RPC
	// use print features function
	// printFeatures(client, &pb.Rectangle{
	// 	Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
	// 	Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	// })

	// client-to-serve streaming RPC
	// RecordRoute
	runRecordRoute(client)

}

func randomPoint(r *rand.Rand) *pb.Point {
	lat := (r.Int31n(180) - 90) * 1e7
	long := (r.Int31n(360) - 180) * 1e7
	return &pb.Point{Latitude: lat, Longitude: long}
}
