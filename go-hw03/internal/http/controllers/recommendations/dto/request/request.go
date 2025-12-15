package request

type ContentToContentRequest struct {
	ContentID string `validate:"required,uuid"`
	Limit     string `validate:"omitempty,numeric"`
	Offset    string `validate:"omitempty,numeric"`
}

type UserContentRequest struct {
	UserID string `validate:"required,uuid"`
	Limit  string `validate:"omitempty,numeric"`
	Offset string `validate:"omitempty,numeric"`
}
