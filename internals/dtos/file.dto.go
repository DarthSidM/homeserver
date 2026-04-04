package dtos

type EditorConfigResponse struct {
	DocumentType string               `json:"documentType"`
	Document     EditorDocumentConfig `json:"document"`
	EditorConfig EditorConfigPayload  `json:"editorConfig"`
}

type EditorDocumentConfig struct {
	FileType string `json:"fileType"`
	Key      string `json:"key"`
	Title    string `json:"title"`
	URL      string `json:"url"`
}

type EditorConfigPayload struct {
	CallbackURL string `json:"callbackUrl"`
	Mode        string `json:"mode"`
}

type OnlyOfficeSavePayload struct {
	Status int    `json:"status"`
	URL    string `json:"url"`
}
