---
id: note-grpc-go
title: '[gRPC] gRPC-GO 筆記'
date: 2020-06-30 10:10:10
updated: 2020-06-30 10:10:10
categories:
  - rpc
keywords:
  - grpc
---

# gRPC-GO 筆記

> - [Basics Tutorial](https://grpc.io/docs/languages/go/basics/) @ gRPC.io
> - Sample Code: [gRPC-go-sandbox](https://github.com/pjchender/grpc-go-sandbox) @ pjchender github
> - [grpc/grpc-go](https://github.com/grpc/grpc-go)
> - [routeguide/server](https://github.com/grpc/grpc-go/blob/master/examples/route_guide/server/server.go) @ Github
> - [routeguide/client](https://github.com/grpc/grpc-go/blob/master/examples/route_guide/client/client.go) @ Github

## TL;DR;

當 protocol buffer 是要在 gRPC 使用的話，需要使用到 `grpc-go` 的套件：

```bash
$ go get -u google.golang.org/grpc
```

或者：

```bash
$ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
$ go get -u google.golang.org/grpc
```

如此就可以使用 `--go_out=plugins=grpc` 的方式，透過 protocol buffer 產生支援 gRPC 的 `.pb.go` 檔：

```bash
# 根據 proto 產生 go 檔
$ protoc --go_out=plugins=grpc:. *.proto
$ protoc -I routeguide/ routeguide/routeguide.proto --go_out=plugins=grpc:routeguide --go_opt=paths=source_relative
```

## 安裝

### compiler

```bash
# 安裝 compiler
$ brew install protobuf
$ protoc --version  # Ensure compiler version is 3+
```

### protobuf

```bash
$ go get -u github.com/golang/protobuf/proto
$ go get -u github.com/golang/protobuf/protoc-gen-go
$ protoc-gen-go --version      # 檢視版本
```

### gRPC-go

當 protocol buffer 是要搭配使用 gRPC 時，需要安裝 `gRPC-go` 的套件：

```bash
$ go get -u google.golang.org/grpc
```

如此才使用`--go_out=plugins=grpc` 的方式，透過 protocol buffer 產生支援 gRPC 的 `.pb.go` 檔。

## Compile 成 pb.go 檔

```bash
# 根據 proto 產生 go 檔
$ protoc --go_out=plugins=grpc:. *.proto

# 產生編譯檔（產生一般的 .pb.go 檔，但沒有使用 gRPC plugin
$ protoc -I=$SRC_DIR --go_out=$DST_DIR $SRC_DIR/addressbook.proto  # 預設根據 go_package 路徑

# 編譯特定的 .proto 檔
# --go_opt=paths=source_relative 表示要用相對路徑產生檔案，如果沒有寫的話，會將檔案根據 module path 建立在指定目錄中
$ protoc --proto_path=src --go_out=build/gen --go_opt=paths=source_relative src/foo.proto src/bar/baz.proto
```

- 使用 `--go_out` 會產生由 `.proto` 編譯後的 `.go` 檔，原本的 `foo.proto` 會變成 `foo.pb.go`
- 在 `.proto` 檔案中，要包含 `go_package` 的屬性，當中要指定包含 generated code 中程式碼的 full Go import path

```go
option go_package = "example.com/foo/bar";
```

- output 資料夾的檔案會根據 `go_package` 的屬性和 compiler flags 有不同

  - 預設情況下，output 檔案會被放在與 `go_package` 屬性相同的資料夾名稱內
  - 透過 `--go_opt=paths=source_relative` 則可以修改 output file 產生的位址

- 如果需要打包成給 gRPC 使用的 GO 檔，需要使用 `--go_out=plugins=grpc:<path>`：

```bash
$ protoc -I grpc/smileinnstatus --go_out=plugins=grpc:grpc/smileinnstatus --go_opt=paths=source_relative grpc/smileinnstatus/smileInnStatus.proto
$ protoc -I routeguide/ routeguide/routeguide.proto --go_out=plugins=grpc:routeguide --go_opt=paths=source_relative
```

## simple RPC

> 範例程式碼：
>
> - [Create Simple RPC Example](https://github.com/pjchender/grpc-go-sandbox/commit/5def9e19b011c61e218439ff6f12b4166dc6d56e) @ Github
> - [simple RPC example(client)](https://github.com/pjchender/grpc-go-sandbox/commit/e592d871100fb27cceafd6dc89bfc0d03435c07d) @ Github

### 1. 使用 proto 定義 service

在 proto 檔中定義 Service：

```protobuf
service RouteGuide {
  // A simple RPC

  // A feature with an empty name is returned if there's no feature at the given position
  rpc GetFeature(Point) returns (Feature) {}
}
```

### 2. 產生 client 和 server 的程式碼

有了 proto 檔後，只需要使用 protocol buffer 的 `protoc` 工具就能夠將自動產生對應程式語言的檔案，例如這裡會是 `routeguide.pb.go`：

```bash
$ protoc -I routeguide/ routeguide/routeguide.proto --go_out=plugins=grpc:routeguide --go_opt=paths=source_relative
```

在這個檔案中將會包含

1. 用來自動產生（populate）、序列化（serialize）和取得 request / response message types 的 protocol buffer 程式碼
2. 給 client 用來使用的 interface type（或稱 stub），以此呼叫定義在 RouteGuide service 中的方法
3. 給 server 用來實作的 interface type

```go
type RouteGuideClient interface {
	// A feature with an empty name is returned if there's no feature at the given position
	GetFeature(ctx context.Context, in *Point, opts ...grpc.CallOption) (*Feature, error)
}

// RouteGuideServer is the server API for RouteGuide service.
type RouteGuideServer interface {
	// A feature with an empty name is returned if there's no feature at the given position
	GetFeature(context.Context, *Point) (*Feature, error)
}
```

### 3. 建立 server

gRPC Server 要做的兩件事：

1. **Implementing Service**：根據 proto 檔中對於 service interface 的定義進行實作，也就是服務真正要做些什麼
2. **Starting the server**：啟動一個 gRPC 伺服器來監聽 clients 發送進來的請求，並且派送到正確的 service 去執行

#### Implementing Service

這段是用來實作 service：

```go
// pb 是 protocol buffer 的簡稱
import pb "sandbox/grpc-go-sandbox/routeguide"

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
	return &pb.Feature{Location: point}, nil
}
```

#### Starting the server

這段是用來啟動 gRPC server：

```go
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
```

### 4. 建立 client

Client 要做得事情包含：

1. Creating the stub：要使用 service 的方法，需要建立一個 gRPC channel 來和 server 溝通。
2. 呼叫 service methods

#### Creating the stub

只需要使用 `grpc.Dial(<serverAddr>)` 即可建立與 server 的 channel：

```go
// STEP 1：creating the client stub
// STEP 1-1：與 gRPC server 建立 channel
// 如果沒有使用安全連線的話，在 options 的地方要加上 grpc.WIthInsecure()
conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
if err != nil {
    log.Fatalf("failed to dial: %v", err)
}
defer conn.Close()
```

一旦 gRPC 的 channel 建立好後，需要 client stub 來執行 RPCs。我們可以使用由 `*.pb.go` 檔（由 proto 檔產生）中，提供 `NewRouteGuideClient` 方法：

```go
import "sandbox/grpc-go-sandbox/routeguide"

// STEP 1-2：使用 proto 所提供的 NewRouteGuideClient 方法並帶入參數 conn 以來建立 client
client := pb.NewRouteGuideClient(conn)
```

#### Calling service methods

在 gPRC-Go 中， RPCs 會以阻塞／同步的方式運算，也就是說一個 RPC 發出去之後，會等待伺服器的回應，不論是的正確的回應或錯誤：

```go
// STEP 2：呼叫 Service Methods
// 透過 context.Context 物件，讓我們在需要時可以改變 RPC 的行為，像是立即執行 time-out/cancel 一個 RPC
feature, err := client.GetFeature(context.Background(), &pb.Point{Latitude: 409146138, Longitude: -746188906})
if err != nil {
    log.Fatalf("failed to getFeature")
}

log.Println(feature)
```

## Server-side streaming RPC

在這個範例中 client 會給一個方形（rectangle），server 則會以 stream 的方式回傳所有在這個 rectangle 中的所有 features。

> 範例程式碼：
>
> - [load features function](https://github.com/pjchender/grpc-go-sandbox/commit/0281aff5032564a6909fef7182591624f8cede94) @ Github
> - [Create server-side streaming gRPC](https://github.com/pjchender/grpc-go-sandbox/commit/c5177d0875bd75f32720b1b9b478265413cf9b4c) @ Github

### 1. 使用 proto 定義 service

在 `service` 中多定義一個 `ListFeatures` 方法，並在 returns 的地方，加上 `stream` 關鍵字，即可讓 server 以串流的方式進行回傳：

```protobuf
// routeguide.proto

service RouteGuide {
  // A server-to-client streaming RPC
  // 結果會以串流的方式回傳，而不是一次傳完
  rpc ListFeatures(Rectangle) returns (stream Feature) {}
}
```

### 2. 產生 client 和 server 的程式碼

這個 protobuffer 的 service 在 build 成 go 的程式碼後，會是可以接收 `Rectangel` 和 `RouteGuide_ListFeaturesServer` 的 method，並且會回傳 error，client 端需要根據這個 error 判斷資料是否已經傳送完畢：

- 當 `error` 為 `nil` 時，表示資料還沒傳完
- 當 `error` 為 `io.EOF` 時，表示資料傳送完畢
- 當 `error` 是其他內容時，表示有錯誤產生

```go
// RouteGuideServer is the server API for RouteGuide service.
type RouteGuideServer interface {
	// A server-to-client streaming RPC
	// 結果會以串流的方式回傳，而不是一次傳完
	ListFeatures(*Rectangle, RouteGuide_ListFeaturesServer) error
}

// RouteGuideClient is the client API for RouteGuide service.
type RouteGuideClient interface {
	ListFeatures(ctx context.Context, in *Rectangle, opts ...grpc.CallOption) (RouteGuide_ListFeaturesClient, error)
}
```

### 3. 建立 server

在 server 端會針對 ListFeatures 這個 method 進行實作：

- 以 `stream.Send()` 的方式將資料以串流回傳

```go
// STEP 1：ListFeatures 會以 server-side stream 的方式將所有 features 回傳給 client
func (s *routeGuideServer) ListFeatures(rect *pb.Rectangle, stream pb.RouteGuide_ListFeaturesServer) error {
	// STEP 2：取出 savedFeatures 的資料並以 stream.Send() 的方式回傳
	// Client 端需要從 err 判斷，如果還有資料未傳完，則 err 會是 nil；如果傳完了會是 io.EOF；否則會得到 err
	for _, feature := range s.savedFeatures {
		if inRange(feature.Location, rect) {
			if err := stream.Send(feature); err != nil {
				return err
			}
		}
	}
	return nil
}
```

### 4. 建立 Client

在 client 端會使用 `client.ListFeatures` 這個方法來取得 server-side streaming 回傳的資料：

- 以 `stream.Recv()` 來取得資料（recv 應該是 receive 的意思）
- 當 `error` 為 `nil` 時，表示資料還沒傳完
- 當 `error` 為 `io.EOF` 時，表示資料傳送完畢
- 當 `error` 是其他內容時，表示有錯誤產生

#### 建立 service methods

```go
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
```

#### 執行該 method

```go
func main() {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRouteGuideClient(conn)

	// Server-side streaming RPC
	// STEP 5: use print features function
	printFeatures(client, &pb.Rectangle{
		Lo: &pb.Point{Latitude: 400000000, Longitude: -750000000},
		Hi: &pb.Point{Latitude: 420000000, Longitude: -730000000},
	})
}
```

## Client-side streaming RPC

### 1. 使用 proto 定義 service

- 定義一個名為 `RecordRoute` 的 service

```protobuf
//
// A client-to-server streaming RPC
service RouteGuide {
  // 以 stream 的方式接收許多 Points，當接收完畢後回傳 RouteSummary
  rpc RecordRoute(stream Point) returns (RouteSummary) {}
}
```

### 2. 產生 client 和 server 的程式碼

- 產生 `RouteSummary struct {}` 來回傳資料，其中包含了 `Reset()`, `String()`

#### 產生 `RouteGuideServer` interface

```go
// RouteGuideServer is the server API for RouteGuide service.
type RouteGuideServer interface {

	// A client-to-server streaming RPC
	// 以 stream 的方式接收許多 Points，當接收完畢後回傳 RouteSummary
	RecordRoute(RouteGuide_RecordRouteServer) error
}


type RouteGuide_RecordRouteServer interface {
	SendAndClose(*RouteSummary) error
	Recv() (*Point, error)
	grpc.ServerStream
}

type routeGuideRecordRouteServer struct {
	grpc.ServerStream
}
func (x *routeGuideRecordRouteServer) SendAndClose(m *RouteSummary) error {/* ... */}
func (x *routeGuideRecordRouteServer) Recv() (*Point, error) {/* ... */}
```

#### 產生 `RouteGuideClient` interface

```go
// RouteGuideClient is the client API for RouteGuide service.
type RouteGuideClient interface {
	// A client-to-server streaming RPC
	// RouteGuide_RecordRouteClient 是 interface
  RecordRoute(ctx context.Context, opts ...grpc.CallOption) (RouteGuide_RecordRouteClient, error)
}

type RouteGuide_RecordRouteClient interface {
	Send(*Point) error
	CloseAndRecv() (*RouteSummary, error)
	grpc.ClientStream
}

type routeGuideRecordRouteClient struct {
	grpc.ClientStream
}
func (x *routeGuideRecordRouteClient) Send(m *Point) error {/* ... */}
func (x *routeGuideRecordRouteClient) CloseAndRecv() (*RouteSummary, error) {/* ... */}
```

### 3. 建立 server

- 透過 `stream.Recv()` 接收 client 傳來的 stream
- 透過 `stream.SendAndClose()` 回傳資料並結束
- 根據 `error` 判斷 stream 是否已經傳完
  - 當 `error` 為 `nil` 時，表示資料還沒傳完
  - 當 `error` 為 `io.EOF` 時，表示資料傳送完畢
  - 當 `error` 是其他內容時，表示有錯誤產生

```go
func (s *routeGuideServer) RecordRoute(stream pb.RouteGuide_RecordRouteServer) error {
	var pointCount, featureCount, distance int32
	var lastPoint *pb.Point
	startTime := time.Now()

	for {
		// 在 server 端接收 client 傳來的 stream
		point, err := stream.Recv()

    // 當 err == io.EOF 時，表示 client 資料傳送完畢
		if err == io.EOF {
			endTime := time.Now()

      // 透過 stream.SendAndClose() 回傳資料給 client
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
```

### 4. 建立 client

#### 建立 service method

- 透過 `RouteGuideClient` 的 `RecordRoute(ctx)` 方法可以取得 `stream`
- 使用 `err := stream.Send(point)` 可以從 client 以 stream 的方式向 server 傳送資料
- 使用 `reply, err := stream.CloseAndRecv()` 可以告知 server 資料傳送完畢，等待接收 server 回應

```go
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
```

#### 執行該 method

```go
func main() {
  // client-to-serve streaming RPC
	// RecordRoute
	runRecordRoute(client)
}
```

## Debugging

> [Debugging](https://github.com/grpc/grpc-go/blob/master/examples/features/debugging/README.md) @ gPRC Github

將環境變數設成：

```bash
GRPC_GO_LOG_VERBOSITY_LEVEL=99
GRPC_GO_LOG_SEVERITY_LEVEL=info
```

## 錯誤排除

### protoc-gen-go: program not found or is not executable

```bash
# 需要把 $GOPATH/bin 加到 .zshrc/.bashrc 等
$ echo 'export PATH=$PATH:$GOPATH/bin' >> $HOME/.zshrc
```
