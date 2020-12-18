package main

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"strings"
)

type Cfg struct {
	ApiKey 		string 		`json:"api_key"`
	UserName 	string 		`json:"user_name"`
	Triggers 	[]string 	`json:"triggers"`
	Chats 		[]string 	`json:"chats"`
	ChatId		int64 		`json:"chat_id"`
}

var config = Cfg{}
var botApi *tgbotapi.BotAPI

func main() {
	authorize()
}

func sendMsg(msg string) {
	message := tgbotapi.NewMessage(config.ChatId, msg)
	_, _ = botApi.Send(message)
}

func authorize(){
	configText, _ := ioutil.ReadFile("config.json")

	if err := json.Unmarshal(configText, &config); err != nil {
		log.Fatal(err)
	}

	authorizeTwitch()
	//authorizeTelegram()
}

func authorizeTwitch() {
	log.Println("auth on twitch")
	client := twitch.NewAnonymousClient()

	client.OnConnect(func() {
		log.Println("connected")
	})
	client.OnPrivateMessage(handleChatMessage)

	client.Join(config.Chats...)
	log.Println("...")
	if err := client.Connect(); err != nil {
		log.Panic(err)
	}
}

func handleChatMessage(message twitch.PrivateMessage){
	// log.Print(message.Message)
	if strings.Contains(strings.ToLower(message.Message), "@" + strings.ToLower(config.UserName)) {
		authorizeTelegram()
		sendMsg(fmt.Sprintf("chat: %s\n@%s: %s", message.Channel, message.User.Name, message.Message))
	} else {
		for _, trigger := range config.Triggers {
			if strings.Contains(strings.ToLower(message.Message), strings.ToLower(trigger)) {
				authorizeTelegram()
				sendMsg(fmt.Sprintf("chat: %s\n@%s: %s", message.Channel, message.User.DisplayName, message.Message))
			}
		}
	}
}

func authorizeTelegram() {
	var err error
	if botApi, err = tgbotapi.NewBotAPI(config.ApiKey); err != nil {
		log.Panic(err)
	}

	botApi.Debug = false

	log.Printf("Authorized on account %s", botApi.Self.UserName)

}