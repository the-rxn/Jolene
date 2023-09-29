package utils

import (
	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

const JHPATH = "./files/joleneHello.ogg"

func SendHelloMessage(chatID int64, bot *tg.BotAPI) {
	helloVoice := tg.FilePath(JHPATH)
	handle, _, err := helloVoice.UploadData()
	if err != nil {
		log.Infof("Couldn't upload file: %s", err)
	}
	log.Debugf("handle: %s", handle)
	msg := tg.NewAudio(chatID, helloVoice)
	bot.Send(msg)
}
