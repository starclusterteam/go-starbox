package scrpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/starclusterteam/go-starbox/log"

	"google.golang.org/grpc"
)

type key int

// Context keys
const (
	_ key = iota
	LOGGERKEY
)

// LoggerInterceptor is a gRPC server-side interceptor that logs requests.
func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	start := time.Now()

	id := generateID()
	ctx = SetLogger(
		ctx,
		GetLogger(ctx).
			With("request_id", id).
			With("method", info.FullMethod),
	)

	resp, err := handler(ctx, req)

	// logger can contain data altered by request
	logger := GetLogger(ctx).
		With("latency", time.Since(start).String())

	if err != nil {
		logger = logger.
			With("error", err).
			With("code", grpc.Code(err))
	}

	logger.Info("request")

	return resp, err
}

// GetLogger returns a logger scoped to request
func GetLogger(ctx context.Context) log.Interface {
	rv := ctx.Value(LOGGERKEY)
	if rv != nil {
		return rv.(log.Interface)
	}
	return log.Logger()
}

// SetLogger sets logger in a context
func SetLogger(ctx context.Context, logger log.Interface) context.Context {
	return context.WithValue(ctx, LOGGERKEY, logger)
}

func generateID() string {
	r := make([]byte, 16)
	_, err := rand.Read(r)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random id: %v", err))
	}

	return hex.EncodeToString(r)
}
