module github.com/safe-waters/retro-simply/backend

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.8.2
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/safe-waters/structured_logger v0.0.0-20210504081827-0cc396eea87e
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/otlp v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/grpc v1.37.0
)
