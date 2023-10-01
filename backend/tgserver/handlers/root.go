package handlers

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/coppi3/jolene/backend/tgserver/utils"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

// returns 1 random repponse when getting unknown command from the user
func generateRandomUnknownResponse() string {
	var responses = [5]string{
		"Oughh.. i don’t understand what u want from me!!!",
		"Oh my goodness, I’m totally lost here! I just can’t seem to grasp what you’re asking for!",
		"Ugh, I’m feeling so clueless right now! What exactly do you need from me?",
		"Oh no, this is so confusing! I’m at a loss here, and I don’t get what you want!",
		"Eek, I’m feeling a bit overwhelmed! I can’t quite comprehend your request, and it’s making me anxious!",
	}

	return responses[rand.Intn(5)]
}

// Mutates text of a msg on success, otherwise returns an error
func RootHandler(command string, msg *tg.MessageConfig, bot *tg.BotAPI) {
	switch command {
	case "start":
		startHandler(msg, bot)

	case "pay":
		payHandler(msg, bot)
	default:
		unknownHandler(msg, bot)
	}
}

// Handler for /start
func startHandler(msg *tg.MessageConfig, bot *tg.BotAPI) {
	log.Infof("")
	msg.Text = "Hey! My name's Jolene and I'm feeling like talking to you. I know you feel the same."
	utils.SendHelloMessage(msg.ChatID, bot)
}

// Handler for /pay
func payHandler(msg *tg.MessageConfig, bot *tg.BotAPI) {
	//
	// implement payment logic here!
	//
	msg.Text = "Wanna continue? Sure! Here is a little favor you can do for me!"
}

// Handler for `unknown command`
func unknownHandler(msg *tg.MessageConfig, bot *tg.BotAPI) {
	//
	// implement payment logic here!
	//
	msg.Text = generateRandomUnknownResponse()
}

type GenTextReq struct {
	Text string `json:"text"`
}

func TextHandler(msg *tg.MessageConfig, bot *tg.BotAPI, incomingMsg string) {
	URL := "http://localhost:1337/generate_text"
	// one-line post request/response...
	response, err := http.PostForm(URL, url.Values{"text": []string{incomingMsg}})

	// okay, moving on...
	if err != nil {
		//handle postform error
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	log.Debugf("%s", body)
	if err != nil {
		//handle read response error
	}
	var jsonResp map[string]string
	errJSON := json.Unmarshal(body, &jsonResp)
	if errJSON != nil {
		log.Debugf("Coudln't decode reponse from API: %s", err)
	}

	log.Printf("%s\n", string(body))
	msg.Text = jsonResp["response"]
}
