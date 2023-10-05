package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/carlmjohnson/requests"

	// "encoding/json"
	// "io"
	"math/rand"
	"net/http"
	"net/url"

	// "net/url"
	"github.com/avast/retry-go/v4"

	"github.com/coppi3/jolene/backend/tgserver/myDb"
	"github.com/coppi3/jolene/backend/tgserver/utils"
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
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

// Handler for a text message
// In the future need to differentiate if incomingMsg is text or voice or photo and split into different headers
func MessageHandler(incomingMsg *tg.Message, bot *tg.BotAPI, db *myDb.DB) {
	log.Infof("[@%s:%d] %s", incomingMsg.From.UserName, incomingMsg.From.ID, incomingMsg.Text)

	dbEntry := myDb.StorageLine{
		UserID:  incomingMsg.From.ID,
		Bot:     false,
		Message: incomingMsg.Text,
		Time:    time.Now(),
	}
	if err := db.Add(dbEntry); err != nil {
		log.Errorf("[DB] Couldn't write to db: %s", err)
	}
	log.Debugln("Inserted 1 incoming message to db")

	// Pretend like we're typing
	chatAction := tg.NewChatAction(incomingMsg.Chat.ID, tg.ChatTyping)
	bot.Send(chatAction)
	// time.Sleep(5 * 1_000) // 5 sec

	msg := tg.NewMessage(incomingMsg.Chat.ID, "")
	rxMsg := incomingMsg.Text
	previousMsgs, err := db.GetMessagesByUserID(incomingMsg.From.ID)
	if err != nil {
		log.Debugf("Encountered error during fetching previous msgs: %s", err)
	}
	promptFromRxMsg := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: rxMsg,
		}}
	if previousMsgs == nil {

		// resp, err := handlers.TextHandler(&msg, bot, rxMsg)
		resp, err := PostGenerateText(promptFromRxMsg)
		if err != nil {
			log.Errorf("Didn't get response from API: %s", err)
		}
		respStorageLine := myDb.StorageLine{
			UserID:  incomingMsg.From.ID,
			Bot:     true,
			Message: resp,
			Time:    time.Now(),
		}
		if err := db.Add(respStorageLine); err != nil {
			log.Errorf("[DB] Couldn't write to db: %s", err)
		}
		log.Debugln("Inserted 1 outgoing message to db")

		bForm := url.Values{
			"text": {resp},
		}
		log.Debugf("Calling API [/generate_voice] with text:%s", bForm["text"][0])
		var voiceLink string
		request := requests.
			URL("http://localhost:1337/generate_voice").
			BodyForm(bForm).
			// Headers(headers).
			ToString(&voiceLink)
		// ErrorJSON(&errorJSON).
		log.Debugf("%q", request)
		err = request.Fetch(context.Background())
		if err != nil {
			log.Infof("Error during API call: %s", err)
		}
		log.Debugf("%s", voiceLink)
		voiceUpload := tg.FileURL(voiceLink)
		if err != nil {
			log.Infof("Couldn't upload file: %s", err)
		}
		voiceMsg := tg.NewAudio(incomingMsg.Chat.ID, voiceUpload)
		bot.Send(voiceMsg)
		if err != nil {
			log.Errorf("Error ocurred during getting voice from API: %s", err)
		}
		msg.Text = resp
	} else {
		var prompts []openai.ChatCompletionMessage
		for _, msg := range previousMsgs {
			if msg.Bot { // if bot
				prompt := openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: msg.Message,
				}
				prompts = append(prompts, prompt)
			} else { // if user
				prompt := openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: msg.Message,
				}
				prompts = append(prompts, prompt)
			}

		}
		prompts = append(prompts, promptFromRxMsg[0])
		log.Debugf("Got %d previous messages from [%s:%d]", len(prompts), incomingMsg.From, incomingMsg.From.ID)
		resp, err := PostGenerateText(prompts)
		if err != nil {
			log.Errorf("Didn't get response from API: %s", err)
		}
		// UserID 0 means it's from bot
		respStorageLine := myDb.StorageLine{
			UserID:  incomingMsg.From.ID,
			Bot:     true,
			Message: resp,
			Time:    time.Now(),
		}
		if err := db.Add(respStorageLine); err != nil {
			log.Errorf("[DB] Couldn't write to db: %s", err)
		}
		log.Debugln("Inserted 1 outgoing message to db")
		msg.Text = resp

	}
	// msg := tg.NewMessage(incomingMsg.Chat.ID, incomingMsg.Text)
	// msg.ReplyToMessageID = incomingMsg.MessageID

	bot.Send(msg)
	err = db.Flush()
	if err != nil {
		log.Errorf("Couldn't flush DB: %s", err)
	}
}

// Mutates text of a msg on success, otherwise returns an error
func RootCommandHandler(command string, msg *tg.MessageConfig, bot *tg.BotAPI) {
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

func PostGenerateText(msgs []openai.ChatCompletionMessage) (string, error) {
	attempts := 3
	log.Println(msgs)
	log.Println(msgs[0].Role, msgs[0].Content)
	resp, err := retry.DoWithData(
		func() (string, error) {
			if attempts == 0 {
				return "", errors.New("Too many unsucessful text generation attempts")
			}
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
					Model:    modelName,
					Messages: msgs,
				},
			)

			if err != nil {
				log.Errorf("ChatCompletion error: %v\n", err)
				return "", err
			}
			respMsg := resp.Choices[0].Message.Content
			if respMsg == "\n" || len(respMsg) < 2 {
				log.Errorf("GOT ERRORLIKE RESPONSE: %s", respMsg)
				attempts--
				return "", errors.New("Too short")
			}
			log.Infof("Got response: %s", resp.Choices[0].Message.Content)
			return resp.Choices[0].Message.Content, nil
		}, retry.RetryIf(func(err error) bool {
			if err.Error() == "Too short" {
				log.Errorf("GOT A TOO SHORT ANSWER, retrting")
				return true
			}
			return false
		}),
	)
	if err != nil {
		log.Errorf("Error during `PostGenerateText()`: %s", err)
	}
	return resp, nil

}
