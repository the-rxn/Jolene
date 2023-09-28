package main

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dot "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

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

	log.Infof("Authorized on account %s", bot.Self.UserName)

	u := tg.NewUpdate(0)
	u.Timeout = 5

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // If it's not a message, we don't care
			continue
		}
		if !update.Message.IsCommand() { // If we get a message, NOT A COMMAND
			log.Infof("[@%s](userid: %d) %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)

			msg := tg.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
		if update.Message.IsCommand() {
			// init an empty message
			msgToSend := tg.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()
			log.Infof("[/%s] from [@%s](userid: %d)\n", command, update.Message.From, update.Message.From.ID)
			handleCommand(command, &msgToSend, bot)
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
		msg.Text = "Oughh.. i don't understand what u want from me!!!"
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
