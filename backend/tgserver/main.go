package main

import (
	"errors"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	dot "github.com/joho/godotenv"
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

const JHPATH = "./files/joleneHello.ogg"

type DB struct {
	sql    *sql.DB
	stmt   *sql.Stmt
	buffer []StorageLine
}
type StorageLine struct {
	userID  int64
	message string
	time    time.Time
}

const (
	insertSQL = `
	INSERT INTO MAIN 
	(userid, message, time) 
	VALUES (?, ?, ?)
	`
	schemaSQL = `
	CREATE TABLE IF NOT EXISTS main (
	userID INTEGER,
	message STRING,
	time TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS main_time ON main(time);
	CREATE INDEX IF NOT EXISTS main_userID ON main(userID);

	`
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

func NewDB(dbFile string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, err
	}
	db := DB{
		sql:    sqlDB,
		stmt:   stmt,
		buffer: make([]StorageLine, 0, 512),
	}
	return &db, nil
}

func (db *DB) Add(storageLine StorageLine) error {
	if len(db.buffer) == cap(db.buffer) {
		return errors.New("storageLines buffer is full")
	}

	db.buffer = append(db.buffer, storageLine)
	if len(db.buffer) == cap(db.buffer) {
		if err := db.Flush(); err != nil {
			log.Error("unable to flush storageLines: %w", err)
			return err
		}
	}

	return nil
}

// Flush pending txs into DB.
func (db *DB) Flush() error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}

	for _, storageLine := range db.buffer {
		_, err := tx.Stmt(db.stmt).Exec(storageLine.userID, storageLine.message, storageLine.time)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	db.buffer = db.buffer[:0]
	return tx.Commit()
}

func (db *DB) Close() error {
	defer func() {
		db.stmt.Close()
		db.sql.Close()
	}()

	if err := db.Flush(); err != nil {
		return err
	}

	return nil
}
func main() {
	// setting up db connection
	dbPath, exists := os.LookupEnv("DBPATH")
	if !exists {
		log.Fatalf("No DB found!")
	}
	db, err := NewDB(dbPath)
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

			dbEntry := StorageLine{
				userID:  update.Message.From.ID,
				message: update.Message.Text,
				time:    time.Now(),
			}
			if err := db.Add(dbEntry); err != nil {
				log.Errorf("[DB] Couldn't write to db: %s", err)
			}
			log.Debugln("Inserted 1 message to db")

			msg := tg.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
		if update.Message.IsCommand() {
			// init an empty message
			msgToSend := tg.NewMessage(update.Message.Chat.ID, "")
			command := update.Message.Command()
			log.Infof("{/%s} [@%s:%d]", command, update.Message.From, update.Message.From.ID)

			// generate response text
			handleCommand(command, &msgToSend, bot)

			//send generated response to the user
			msgToSend.ReplyToMessageID = update.Message.MessageID
			bot.Send(msgToSend)
		}
		db.Flush()
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
		log.Infof("Couldn't upload file: %s", err)
	}
	log.Debugf("handle: %s", handle)
	msg := tg.NewAudio(chatID, helloVoice)
	bot.Send(msg)
}
