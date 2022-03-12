package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	speech "cloud.google.com/go/speech/apiv1"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Panic("Missing telegram token")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.Voice != nil { // If we got a message
			go func(update tgbotapi.Update) {
				log.Printf("[%s]", update.Message.From.UserName)

				url, err := bot.GetFileDirectURL(update.Message.Voice.FileID)
				if err != nil {
					log.Println(err)
				}
				file, err := DownloadFile(url)
				if err != nil {
					log.Println(err)
					return
				}

				text, err := SpeechToText(file)
				if err != nil {
					log.Println(err)
					return
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
				msg.ReplyToMessageID = update.Message.MessageID
				msg.Text = text

				_, err = bot.Send(msg)
				if err != nil {
					log.Println(err)
					return
				}
			}(update)
		}
	}
}

func SpeechToText(file []byte) (string, error) {
	ctx := context.Background()

	// Creates a client.
	client, err := speech.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Detects speech in the audio file.
	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_OGG_OPUS,
			SampleRateHertz: 48000,
			LanguageCode:    "ru-RU",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{
				Content: file,
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to recognize: %v", err)
	}

	// // Prints the results.
	// for _, result := range resp.Results {
	// 	fmt.Printf("----- %+v", result)
	// 	for _, alt := range result.Alternatives {
	// 		fmt.Printf("\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
	// 	}
	// }

	// Prints billed time.
	log.Println("Billed time:", resp.TotalBilledTime)

	textResult := strings.Builder{}
	for _, result := range resp.Results {
		textResult.WriteString(result.Alternatives[0].Transcript)
	}

	return textResult.String(), nil
}

func DownloadFile(URL string) ([]byte, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buffer := bytes.NewBuffer([]byte{})

	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
