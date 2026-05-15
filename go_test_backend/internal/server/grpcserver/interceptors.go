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
	// 给刚接触 Go 的同学：
	// - Go 的函数可以返回另一个函数（闭包 closure）。
	// - grpc.UnaryServerInterceptor 是 gRPC 定义的一种函数类型，用来“拦截/包装”请求处理流程。
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
	// 给刚接触 Go 的同学：
	// - `any` 等价于 interface{}，表示“任意类型”。
	// - reflect（反射）允许我们在运行时动态查看/读取一个值的结构（这里用来找 UserID 字段）。
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
