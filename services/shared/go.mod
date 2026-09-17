module github.com/pawradise/shared

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.18.0
	google.golang.org/grpc v1.61.0
	google.golang.org/protobuf v1.32.0
)

// Note: redis is an optional dependency — use go get to add when needed.
