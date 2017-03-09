package models

type FileUploadResponse struct {
	Name   string `json:"name"`
	UUID   string `json:"uuid"`
	Md5    string `json:"md5"`
	Sha256 string `json:"sha256"`
}
