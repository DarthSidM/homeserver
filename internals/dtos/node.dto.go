package dtos

type RenameNodeRequest struct {
	NewName string `json:"new_name" validate:"required,min=1,max=255"`
}

type RenameNodeResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Message string `json:"message"`
}
