package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coppi3/jolene/backend/tgserver/handlers"
	"github.com/coppi3/jolene/backend/tgserver/myDb"

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
		if !update.Message.IsCommand() {
			handlers.MessageHandler(update.Message, bot, db)
		}
		if update.Message.IsCommand() {
			// init an empty message
			msgToSend := tg.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()
			log.Infof("{/%s} [@%s:%d]", command, update.Message.From, update.Message.From.ID)

			// generate response text
			handlers.RootCommandHandler(command, &msgToSend, bot)

			//send generated response to the user
			msgToSend.ReplyToMessageID = update.Message.MessageID
			bot.Send(msgToSend)
		}
	}
}
