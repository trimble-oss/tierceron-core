module github.com/trimble-oss/tierceron-core/v2

go 1.26.2

require (
	github.com/glycerine/bchan v0.0.0-20170210221909-ad30cd867e1c
	github.com/go-git/go-billy/v5 v5.6.2
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/trimble-oss/tierceron-nute-core v1.0.4
	golang.org/x/sys v0.41.0
	google.golang.org/grpc v1.79.3
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/stretchr/testify v1.11.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
)

replace (
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.43.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.43.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.43.0
)
