package user

// User 是 HTTP API 与 gRPC 之间共用的数据结构。
// JSON tag（例如 `json:"user_id"`）用来控制 JSON 字段名。
type User struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Mock   bool   `json:"mock"`
}

// QueryRequest 是“按 user_id 查询”的请求参数。
type QueryRequest struct {
	UserID string `json:"user_id"`
}
