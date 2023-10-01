package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
)

type GenerateTextResp struct {
	Response string `json:"response"`
}

func GetRoot(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func PostGenerateText(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
	// reading body
	text := r.PostFormValue("text")
	fmt.Println(text)

	// chat stuff
	config := openai.ClientConfig{
		BaseURL:    "http://localhost:1489/",
		HTTPClient: http.DefaultClient,
	}
	client := openai.NewClientWithConfig(config)
	modelName := "LLaMA_CPP"
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: text,
				},
			},
		},
	)

	if err != nil {
		error := fmt.Sprintf("ChatCompletion error: %v\n", err)
		w.Header().Set("Error", error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := GenerateTextResp{Response: resp.Choices[0].Message.Content}
	// responseJSON, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Error", "Couldn't encode response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
	// io.WriteString(w, responseJSON)
}
func PostGenerateVoice(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func PostConvertToAudio(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
