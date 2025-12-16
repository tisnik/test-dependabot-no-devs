package domain

type FailResponse struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

type ErrorResponse struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Code    int               `json:"code"`
	Data    map[string]string `json:"data"`
}
