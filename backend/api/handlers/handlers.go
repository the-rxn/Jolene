package handlers

import (
	"io"
	"net/http"
)

func GetRoot(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func PutGenerateText(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func PutGenerateVoice(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func PutConvertToAudio(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
