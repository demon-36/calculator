package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"calculator/calculator"
)

const MaxBodyBytes = 16 << 10

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateSuccessResponse struct {
	Result float64 `json:"result"`
}

type CalculateErrorResponse struct {
	Error *calculator.Problem `json:"error"`
}

func WriteProblem(w http.ResponseWriter, p *calculator.Problem) {
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(CalculateErrorResponse{Error: p})
}

func WriteSuccess(w http.ResponseWriter, result float64) {
	_ = json.NewEncoder(w).Encode(CalculateSuccessResponse{Result: result})
}

func ServeAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path != "/api/calculate" {
		WriteProblem(w, &calculator.Problem{Status: 404, Code: "not_found", Message: "Unknown endpoint."})
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		WriteProblem(w, &calculator.Problem{Status: 405, Code: "method_not_allowed", Message: "Use POST for calculations."})
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		WriteProblem(w, &calculator.Problem{Status: 415, Code: "unsupported_media_type", Message: "Send application/json."})
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
	if err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			WriteProblem(w, &calculator.Problem{Status: 413, Code: "body_too_large", Message: "Request exceeds 16 KiB."})
		} else {
			WriteProblem(w, &calculator.Problem{Status: 400, Code: "invalid_json", Message: "Could not read the request."})
		}
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var request CalculateRequest
	if err := decoder.Decode(&request); err != nil || decoder.Decode(new(json.RawMessage)) != io.EOF {
		WriteProblem(w, &calculator.Problem{Status: 400, Code: "invalid_json", Message: "Send one JSON object with an expression string."})
		return
	}
	result, p := calculator.Evaluate(request.Expression)
	if p != nil {
		WriteProblem(w, p)
		return
	}
	WriteSuccess(w, result)
}

func NewRouter() http.Handler {
	return http.HandlerFunc(ServeAPI)
}
