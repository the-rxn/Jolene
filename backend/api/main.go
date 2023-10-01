package main

import (
	"github.com/coppi3/jolene/backend/api/handlers"
	"net/http"
	"os"

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

// these are endpoints:
// 	1. /generate_text(text) -> text
//	2. /generate_voice(text) -> id of a voice
//	3. /convert_to_audio(id) -> file

func main() {
	// get port
	SERVERPORT, exists := os.LookupEnv("SERVERPORT")
	if !exists {
		log.Fatalf("No SERVERPORT in .env found!")
	}
	// init multiplexer and handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.GetRoot)
	mux.HandleFunc("/generate_text", handlers.PostGenerateText)
	mux.HandleFunc("/generate_voice", handlers.PostGenerateVoice)
	mux.HandleFunc("/convert_to_audio", handlers.PostConvertToAudio)

	err := http.ListenAndServe(SERVERPORT, mux)
	if err != nil {
		log.Fatalf("Error ocurred while serving: %s", err)
	}
}
