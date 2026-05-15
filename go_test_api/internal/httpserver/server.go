package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go_test/shared/user"
	"go_test/shared/usergrpc"
)

type Server struct {
	log     *zap.Logger
	userCli usergrpc.UserServiceClient
	timeout time.Duration
}

// New 创建一个 HTTP Server，它内部会通过 gRPC 调用后端。
func New(log *zap.Logger, userCli usergrpc.UserServiceClient, timeout time.Duration) *Server {
	return &Server{log: log, userCli: userCli, timeout: timeout}
}

// Handler 返回一个 http.Handler，并注册这些路由：
// - POST   {basePath}/users      新增用户
// - GET    {basePath}/users/{id} 查询用户
// - GET    {basePath}/health     健康检查
func (s *Server) Handler(basePath string) http.Handler {
	mux := http.NewServeMux()
	prefix := strings.TrimRight(basePath, "/")

	mux.HandleFunc(prefix+"/users", func(w http.ResponseWriter, r *http.Request) { s.handleUsers(w, r) })
	mux.HandleFunc(prefix+"/users/", func(w http.ResponseWriter, r *http.Request) { s.handleUserByID(w, r, prefix) })
	mux.HandleFunc(prefix+"/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	return mux
}

// handleUsers 处理 POST /users：创建用户。
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	var in user.User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	s.log.Info("http request", zap.String("path", r.URL.Path), zap.String("user_id", in.UserID))

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	out, err := s.userCli.Add(ctx, &in)
	if err != nil {
		writeJSON(w, httpStatusFromError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleUserByID 处理 GET /users/{id}：按 ID 查询用户。
func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request, prefix string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}

	// URL 形如：/users/{id}
	want := prefix + "/users/"
	if !strings.HasPrefix(r.URL.Path, want) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, want)
	if idStr == "" || strings.Contains(idStr, "/") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid user_id"})
		return
	}
	s.log.Info("http request", zap.String("path", r.URL.Path), zap.String("user_id", idStr))

	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	out, err := s.userCli.Query(ctx, &user.QueryRequest{UserID: idStr})
	if err != nil {
		writeJSON(w, httpStatusFromError(err), map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// writeJSON 输出 JSON 响应，并设置 HTTP 状态码。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
