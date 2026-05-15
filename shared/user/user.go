package user

type User struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type QueryRequest struct {
	UserID string `json:"user_id"`
}
