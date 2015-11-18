package bosh

import (
	"encoding/json"
	"fmt"
)

const (
	DefaultErrorType      = "Bosh::Clouds::CloudError"
	NotSupportedErrorType = "Bosh::Clouds::NotSupported"
)

type ResponseError struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Retryable bool   `json:"ok_to_retry"`
}

type CpiResponse struct {
	Result interface{}    `json:"result"`
	Error  *ResponseError `json:"error"`
	Log    string         `json:"log"`
}

func BuildDefaultErrorResponse(err error, retryable bool, logOutput string) string {
	return BuildErrorResponse(err, DefaultErrorType, retryable, logOutput)
}

func BuildErrorResponse(err error, errType string, retryable bool, logOutput string) string {
	res := CpiResponse{Log: logOutput}

	resErr := ResponseError{
		Type:      errType,
		Message:   err.Error(),
		Retryable: retryable,
	}
	res.Error = &resErr

	resBytes, err := json.Marshal(res)
	if err != nil {
		return fmt.Sprint(`{"result": null, "error": {"type": "i/o", "message": "marshalling error response", "ok_to_retry": false}, "log": ""}`)
	}

	return fmt.Sprint(string(resBytes))
}

func BuildResultResponse(result interface{}, logOutput string) string {
	res := CpiResponse{Result: result, Log: logOutput}

	resBytes, err := json.Marshal(res)
	if err != nil {
		return fmt.Sprint(`{"result": null, "error": {"type": "i/o", "message": "marshalling success response", "ok_to_retry": false}, "log": ""}`)
	}

	return fmt.Sprint(string(resBytes))
}
