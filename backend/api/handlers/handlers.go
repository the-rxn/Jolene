package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
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
	// resp, err := client.Create
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
	log.Infof("Got response: %s", resp)
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

type GenerateVoiceRequest struct {
	Text  string `json:"text"`
	Voice string `json:"voice"`

	Quality string `json:"quality"`

	OutputFormat string `json:"output_format"`

	Speed int `json:"speed"`

	SampleRate int `json:"sample_rate"`
}

type GenerateVoiceResponse struct {
}

func PostGenerateVoice(w http.ResponseWriter, r *http.Request) {

	text := r.PostFormValue("text")
	log.Debugf("[/generate_voice] got formValue Text : %s", text)
	// url := "https://play.ht/api/v2/tts"
	payload := GenerateVoiceRequest{
		Text:         text,
		Voice:        "s3://voice-cloning-zero-shot/39ae6d35-1ee9-4bfd-8d61-85868c5bb98b/seductive1/manifest.json",
		Quality:      "medium",
		OutputFormat: "ogg",
		Speed:        1,
		SampleRate:   24000,
	}
	headers := map[string][]string{
		// "accept":        {"text/event-stream"},
		"accept":        {"application/json"},
		"content-type":  {"application/json"},
		"AUTHORIZATION": {"286aa3a9d3964acbb77853e0d79fd934"},
		"X-USER-ID":     {"NDvxEXyoSgb8iCvORo4aYvZUGv52"},
	}
	var respBuf bytes.Buffer
	var errorJSON struct {
		Error_message string `json:"error_message`
		Error_id      string `json:"error_id"`
	}
	request := requests.
		URL("/api/v2/tts").
		Host("play.ht").
		BodyJSON(&payload).
		Headers(headers).
		ToBytesBuffer(&respBuf).
		ErrorJSON(&errorJSON).
		Post()
	log.Debugf("%q", request)
	err := request.Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to play.ht: ", err)
		fmt.Println(errorJSON)
	}
	var respJson map[string]interface{}
	err = json.Unmarshal(respBuf.Bytes(), &respJson)
	if err != nil {
		fmt.Println("could not connect deserialize response ", err)
	}

	log.Debugln(respJson)
	id := respJson["id"]
	var voice_url string
	// go func()
	for {
		status, link, err := getDoneVoice(id.(string))
		if err != nil {
			log.Errorf("Error during waiting for voice")
		}
		if status {
			voice_url = link
			log.Debugf("Got link! %s", voice_url)
			break
		}
		time.Sleep(time.Second * 5)
		continue
	}
	// }()

	log.Debugln(id)

	// ctx := r.Context()
	io.WriteString(w, voice_url)
}
func PostConvertToAudio(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	io.WriteString(w, "This is an API for Jolene. If you don't know how you got here, please, contact the owner @hdydylmaily using telegram.")
}
func getDoneVoice(id string) (bool, string, error) {
	log.Debugf("`getDoneVoice()` is being ran: %s", time.Now().Local())
	headers := map[string][]string{
		// "accept":        {"text/event-stream"},
		"accept":        {"application/json"},
		"AUTHORIZATION": {"286aa3a9d3964acbb77853e0d79fd934"},
		"X-USER-ID":     {"NDvxEXyoSgb8iCvORo4aYvZUGv52"},
	}
	var respBuf bytes.Buffer
	var errorJSON struct {
		Error_message string `json:"error_message`
		Error_id      string `json:"error_id"`
	}
	url := fmt.Sprintf("/api/v2/tts/%s", id)
	request := requests.
		URL(url).
		Host("play.ht").
		Headers(headers).
		ToBytesBuffer(&respBuf).
		ErrorJSON(&errorJSON)
	log.Debugf("%q", request)
	log.Debugln(respBuf)
	err := request.Fetch(context.Background())

	if err != nil {
		fmt.Println("could not connect to play.ht: ", err)
		fmt.Println(errorJSON)
	}

	var respJson map[string]interface{}
	if respBuf.Len() != 0 {
		err = json.Unmarshal(respBuf.Bytes(), &respJson)
		if err != nil {
			fmt.Println("could not connect deserialize response ", err)
		}
		log.Debugf("%q", respJson)
		output := respJson["output"]
		if output == nil {
			return false, "", nil
		}
		link := output.(map[string]any)["url"].(string)
		return true, link, nil
	}
	return false, "", nil

}

// def generate_audio_wav_link(text: str) -> str:
//     url = "https://play.ht/api/v2/tts"
//     payload = {
//         "text": f"{text}",
//         "voice": "s3://voice-cloning-zero-shot/39ae6d35-1ee9-4bfd-8d61-85868c5bb98b/seductive1/manifest.json",
//         "quality": "medium",
//         "output_format": "wav",
//         "speed": 1,
//         "sample_rate": 24000
//     }
//     headers = {
//         "accept": "text/event-stream",
//         "content-type": "application/json",
//         "AUTHORIZATION": "286aa3a9d3964acbb77853e0d79fd934",
//         "X-USER-ID": "NDvxEXyoSgb8iCvORo4aYvZUGv52"
//     }
//
//     response = requests.post(url, json=payload, headers=headers)
//     data = response.text
//     data_with_no_newline = "".join([s for s in data.splitlines(True) if s.strip()])
//     print(data_with_no_newline)
//     jsoned = json.loads(data_with_no_newline.splitlines()[-1].replace("data: ", ""))
//     print(jsoned)
//     return jsoned.get("url")
