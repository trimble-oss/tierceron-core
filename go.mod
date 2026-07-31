module github.com/trimble-oss/tierceron-core/v2

go 1.26.4

require (
	github.com/glycerine/bchan v0.0.0-20170210221909-ad30cd867e1c
	github.com/go-git/go-billy/v5 v5.9.0
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/trimble-oss/tierceron-nute-core v1.0.7
	golang.org/x/sys v0.47.0
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
)

replace (
	go.opentelemetry.io/otel => go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/metric => go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/sdk => go.opentelemetry.io/otel/sdk v1.43.0
	go.opentelemetry.io/otel/sdk/metric => go.opentelemetry.io/otel/sdk/metric v1.43.0
	go.opentelemetry.io/otel/trace => go.opentelemetry.io/otel/trace v1.43.0
)
