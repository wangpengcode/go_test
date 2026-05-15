package httpserver

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func httpStatusFromError(err error) int {
	// 给刚接触 Go 的同学：
	// - gRPC 有自己的状态码（codes.InvalidArgument、codes.NotFound 等）。
	// - 这里把 gRPC 状态码映射成 HTTP 状态码，方便 REST 客户端理解。
	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusBadGateway
	}
}
