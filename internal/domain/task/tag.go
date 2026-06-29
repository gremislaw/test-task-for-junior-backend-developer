package task

type Tag struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    IsSystem bool   `json:"is_system"`
}