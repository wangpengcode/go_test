package grpcclient

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func unaryClientLogger(log *zap.Logger) grpc.UnaryClientInterceptor {
	// UnaryClientInterceptor 是一种函数类型，可以“包装”每一次发出去的 gRPC 调用。
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		if userID, ok := extractUserID(req); ok {
			log.Info("grpc client request", zap.String("method", method), zap.String("user_id", userID))
		} else {
			log.Info("grpc client request", zap.String("method", method))
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		st := status.Convert(err)
		if err != nil {
			fields := []zap.Field{zap.String("method", method), zap.Int("code", int(st.Code())), zap.Duration("duration", time.Since(start)), zap.String("error", err.Error())}
			if userID, ok := extractUserID(req); ok {
				fields = append(fields, zap.String("user_id", userID))
			}
			log.Warn("grpc client response", fields...)
			return err
		}
		fields := []zap.Field{zap.String("method", method), zap.Int("code", int(st.Code())), zap.Duration("duration", time.Since(start))}
		if userID, ok := extractUserID(req); ok {
			fields = append(fields, zap.String("user_id", userID))
		}
		log.Info("grpc client response", fields...)
		return nil
	}
}

func extractUserID(v any) (string, bool) {
	// 这里用反射尝试读取请求里的 "UserID"/"UserId" 字段，用于日志追踪。
	// 好处：日志逻辑更通用，不要求每种请求类型都实现某个特定接口。
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
