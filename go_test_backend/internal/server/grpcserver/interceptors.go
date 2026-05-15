package grpcserver

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func UnaryServerLogger(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		userID, ok := extractUserID(req)
		if ok {
			log.Info("grpc request", zap.String("method", info.FullMethod), zap.String("user_id", userID))
		} else {
			log.Info("grpc request", zap.String("method", info.FullMethod))
		}

		resp, err = handler(ctx, req)
		st := status.Convert(err)
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Int("code", int(st.Code())),
			zap.Duration("duration", time.Since(start)),
		}
		if ok {
			fields = append(fields, zap.String("user_id", userID))
		}
		if err != nil {
			fields = append(fields, zap.String("error", err.Error()))
			log.Warn("grpc response", fields...)
			return resp, err
		}
		log.Info("grpc response", fields...)
		return resp, nil
	}
}

func extractUserID(v any) (string, bool) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return "", false
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "", false
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return "", false
	}
	f := rv.FieldByName("UserID")
	if !f.IsValid() {
		f = rv.FieldByName("UserId")
	}
	if !f.IsValid() {
		return "", false
	}
	switch f.Kind() {
	case reflect.String:
		return f.String(), f.String() != ""
	case reflect.Int64, reflect.Int, reflect.Int32:
		return fmt.Sprintf("%d", f.Int()), f.Int() != 0
	default:
		return "", false
	}
}
