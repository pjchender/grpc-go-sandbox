# GRPC-GO-Sandbox

> - [Basics Tutorial](https://grpc.io/docs/languages/go/basics/) @ gRPC.io
> - [gRPC-go-sandbox](https://github.com/pjchender/grpc-go-sandbox) @ pjchender github

## simple RPC

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

### Creating the client

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
    log.Fatalf("faild to dial: %v", err)
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





