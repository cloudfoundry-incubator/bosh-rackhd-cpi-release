package cpi

type FileMetadataResponse []struct {
	Basename string `json:"basename"`
	Filename string `json:"filename"`
	UUID string `json:"uuid"`
	Md5 string `json:"md5"`
	Sha256 string `json:"sha256"`
	Version int `json:"version"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}