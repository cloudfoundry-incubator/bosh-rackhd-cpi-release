package models

type Task struct {
	Name           string                 `json:"injectableName"`
	UnusedName     string                 `json:"friendlyName"`
	ImplementsTask string                 `json:"implementsTask"`
	Options        map[string]interface{} `json:"options"`
	Properties     TaskProperties         `json:"properties"`
}

type TaskProperties struct{}
