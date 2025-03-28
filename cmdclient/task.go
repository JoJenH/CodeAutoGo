package cmdclient

type Task struct {
	Status   string  `json:"status"`
	Progress float32 `json:"progress"`
	Error    string  `json:"error,omitempty"`
}
