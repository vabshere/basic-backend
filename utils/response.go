package utils

import (
	"encoding/json"
	"net/http"
)

// Response is the type for reponses with a simple string message
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResponseStruct is the type for responses with struct data
type ResponseStruct struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// Respond sends a JSON response of type Response
func Respond(code int, msg string, statusCode int, w http.ResponseWriter, r *http.Request) {
	res := Response{code, msg}
	resJson, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resJson)
	return
}

// RespondJson sends a JSON response of type ResponseStruct
func RespondJson(code int, data interface{}, statusCode int, w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{code, data}
	resJson, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(resJson)
	return
}
