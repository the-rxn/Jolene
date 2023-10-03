package main

import (
	// "fmt"
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/coppi3/jolene/backend/tgserver/handlers"
	"github.com/coppi3/jolene/backend/tgserver/myDb"
	"github.com/sashabaranov/go-openai"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dot "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {
	// setting up logus
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{DisableLevelTruncation: true, ForceColors: true})
	// setting up dotenv
	if err := dot.Load(); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

}

func main() {
	// setting up db connection
	dbPath, exists := os.LookupEnv("DBPATH")
	if !exists {
		log.Fatalf("No DB found!")
	}
	db, err := myDb.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Coudln't connect to the database: %s", err)
	}
	defer func() {
		log.Infof("Closing db connection.")
		db.Close()
	}()

	// setting up telegram api
	tgApiKey, exists := os.LookupEnv("TGAPIKEY")
	if exists {
		log.Debugf("TGAPIKEY: %s", tgApiKey)
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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-shutdown
		signal.Stop(shutdown)
		log.Printf("Got a %s signal, stop receiving updates, app is closing...", s)
		db.Close()
		bot.StopReceivingUpdates()
	}()

	for update := range updates {
		if update.Message == nil { // If it's not a message, we don't care
			continue
		}
		if !update.Message.IsCommand() { // If we get a message, NOT A COMMAND
			log.Infof("[@%s:%d] %s", update.Message.From.UserName, update.Message.From.ID, update.Message.Text)

			dbEntry := myDb.StorageLine{
				UserID:  update.Message.From.ID,
				Bot:     false,
				Message: update.Message.Text,
				Time:    time.Now(),
			}
			if err := db.Add(dbEntry); err != nil {
				log.Errorf("[DB] Couldn't write to db: %s", err)
			}
			log.Debugln("Inserted 1 incoming message to db")

			// Pretend like we're typing
			chatAction := tg.NewChatAction(update.Message.Chat.ID, tg.ChatTyping)
			bot.Send(chatAction)
			// time.Sleep(5 * 1_000) // 5 sec

			msg := tg.NewMessage(update.Message.Chat.ID, "")
			rxMsg := update.Message.Text
			// rxMsg = fmt.Sprintf("[INST]%s[/INST]", rxMsg)
			previousMsgs, err := db.GetMessagesByUserID(update.Message.From.ID)
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
				resp, err := handlers.PostGenerateText(promptFromRxMsg)
				if err != nil {
					log.Errorf("Didn't get response from API: %s", err)
				}
				respStorageLine := myDb.StorageLine{
					UserID:  update.Message.From.ID,
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
				voiceMsg := tg.NewAudio(update.Message.Chat.ID, voiceUpload)
				bot.Send(voiceMsg)
				if err != nil {
					log.Errorf("Error ocurred during getting voice from API: %s", err)
				}
				msg.Text = resp
			} else {
				var prompts []openai.ChatCompletionMessage
				for _, msg := range previousMsgs {
					if msg.Bot { // if bot
						// prompt = fmt.Sprintf("%s[INST]%s[/INST]\n", prompt, msg.Message)
						prompt := openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleAssistant,
							Content: msg.Message,
						}
						prompts = append(prompts, prompt)
					} else { // if user
						// prompt = fmt.Sprintf("%s\n%s\n", prompt, msg.Message)
						prompt := openai.ChatCompletionMessage{
							Role:    openai.ChatMessageRoleUser,
							Content: msg.Message,
						}
						prompts = append(prompts, prompt)
					}

				}
				prompts = append(prompts, promptFromRxMsg[0])
				// prompt = fmt.Sprintf("%s%s", prompt, rxMsg)
				// log.Debugf("Got this previous messages from [%s:%d]\n---\n%s\n---", update.Message.From, update.Message.From.ID, prompts)
				log.Debugf("Got %d previous messages from [%s:%d]", len(prompts), update.Message.From, update.Message.From.ID)
				// resp, err := handlers.TextHandler(&msg, bot, prompts)
				resp, err := handlers.PostGenerateText(prompts)
				if err != nil {
					log.Errorf("Didn't get response from API: %s", err)
				}
				// UserID 0 means it's from bot
				respStorageLine := myDb.StorageLine{
					UserID:  update.Message.From.ID,
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
			// msg := tg.NewMessage(update.Message.Chat.ID, update.Message.Text)
			// msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
			err = db.Flush()
			if err != nil {
				log.Errorf("Couldn't flush DB: %s", err)
			}
		}
		if update.Message.IsCommand() {
			// init an empty message
			msgToSend := tg.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()
			log.Infof("{/%s} [@%s:%d]", command, update.Message.From, update.Message.From.ID)

			// generate response text
			handlers.RootHandler(command, &msgToSend, bot)

			//send generated response to the user
			msgToSend.ReplyToMessageID = update.Message.MessageID
			bot.Send(msgToSend)
		}
	}
}
