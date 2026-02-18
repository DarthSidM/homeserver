package dtos

import "github.com/google/uuid"

type CreateDirectoryRequest struct {
	Name     string     `json:"name" validate:"required,min=1,max=255"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

type CreateDirectoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	Type      string    `json:"type"` // always "directory"
	CreatedAt int64     `json:"created_at"`
}
