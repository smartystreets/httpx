package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

func WriteResult(response http.ResponseWriter, statusCode int) {
	WriteErrorMessage(response, http.StatusText(statusCode), statusCode)
}

func WriteResponse(response http.ResponseWriter, err error) {
	if err != nil {
		WriteError(response, err, http.StatusInternalServerError)
	} else {
		response.Header().Set(ContentTypeHeader, MIMEApplicationJSON)
	}
}

func WriteError(response http.ResponseWriter, err error, statusCode int) {
	WriteErrorMessage(response, err.Error(), statusCode)
}

func WriteErrorMessage(response http.ResponseWriter, message string, statusCode int) {
	http.Error(response, message, statusCode)
}

func WriteRequest(response http.ResponseWriter, request *http.Request, message string, status int) {
	dump, _ := httputil.DumpRequest(request, false)
	http.Error(response, fmt.Sprintf("%d %s\n\nRaw Request:\n\n%s", status, message, string(dump)), status)
}

func WriteJSON(contents interface{}, response http.ResponseWriter) {
	response.Header().Set(ContentTypeHeader, MIMEApplicationJSON)
	json.NewEncoder(response).Encode(contents)
}

func WritePrettyJSON(contents interface{}, response http.ResponseWriter) {
	response.Header().Set(ContentTypeHeader, MIMEApplicationJSON)
	encoder := json.NewEncoder(response)
	encoder.SetIndent("", "\t")
	encoder.Encode(contents)
}

const (
	ContentTypeHeader   = "Content-Type"
	MIMEApplicationJSON = "application/json; charset=utf-8"
	MIMETextPlain       = "text/plain; charset=utf-8"
)
