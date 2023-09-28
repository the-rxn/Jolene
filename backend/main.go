package main

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dot "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
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

const JHPATH = "./files/joleneHello.ogg"

func init() {
	// setting up logus
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableLevelTruncation: true, ForceColors: true})
	// setting up dotenv
	if err := dot.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s\n", err)
	}
}
func main() {
	tgApiKey, exists := os.LookupEnv("TGAPIKEY")
	if exists {
		log.Debugf("TGAPIKEY: %s\n", tgApiKey)
	}
	bot, err := tg.NewBotAPI(tgApiKey)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true
	bot.Debug = false

	log.Infof("Authorized on account @%s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 5

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // If it's not a message, we don't care
			continue
		}
		if !update.Message.IsCommand() { // If we get a message, NOT A COMMAND
			log.Infof("[@%s:%d] %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)

			msg := tg.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
		if update.Message.IsCommand() {
			// init an empty message
			msgToSend := tg.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()
			log.Infof("{/%s} [@%s:%d]\n", command, update.Message.From, update.Message.From.ID)

			// generate response text
			handleCommand(command, &msgToSend, bot)

			//send generated response to the user
			msgToSend.ReplyToMessageID = update.Message.MessageID
			bot.Send(msgToSend)
		}
	}
}

// Mutates text of a message on sucess, otherwise returns an error
func handleCommand(command string, msg *tg.MessageConfig, bot *tg.BotAPI) {
	switch command {
	case "start":
		log.Infof("")
		msg.Text = "Hey! My name's Jolene and I'm feeling like talking to you. I know you feel the same."
		sendHelloMessage(msg.ChatID, bot)
	case "pay":
		msg.Text = "Wanna continue? Sure! Here is a little favor you can do for me!"
		//
		// implement payment logic here!
		//
	default:
		msg.Text = generateRandomUnknownResponse()
	}
}

func sendHelloMessage(chatID int64, bot *tg.BotAPI) {
	helloVoice := tg.FilePath(JHPATH)
	handle, _, err := helloVoice.UploadData()
	if err != nil {
		log.Infof("Couldn't upload file: %s\n", err)
	}
	log.Debugf("handle: %s\n", handle)
	msg := tg.NewAudio(chatID, helloVoice)
	bot.Send(msg)
}
