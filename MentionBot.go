package main

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

type Cfg struct {
	ApiKey   string   `json:"api_key"`
	UserName string   `json:"user_name"`
	Triggers []string `json:"triggers"`
	Chats    []string `json:"chats"`
	ChatId   int64    `json:"chat_id"`
}

var config = Cfg{}
var botApi *tgbotapi.BotAPI

func main() {
	authorize()
}

func sendMsg(msg string) {
	message := tgbotapi.NewMessage(config.ChatId, msg)
	message.ParseMode = "Markdown"
	_, _ = botApi.Send(message)
}

func authorize() {
	configText, _ := ioutil.ReadFile("config.json")

	if err := json.Unmarshal(configText, &config); err != nil {
		log.Fatal(err)
	}

	if config.ChatId == 0 {
		authorizeTelegram()

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates, _ := botApi.GetUpdatesChan(u)
		println("Write something to your bot to get chat ID")
		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}

			log.Printf("your ID:%d\n", update.Message.Chat.ID)
			config.ChatId = update.Message.Chat.ID
			break
		}
	}

	authorizeTwitch()
	//authorizeTelegram()
}

func removeNicknames(s string) string {
	var re = regexp.MustCompile(`@\w+`)
	result := ""
	final := ""
	var matches = re.FindAllStringSubmatchIndex(s, -1)
	last := 0
	for _, indices := range matches {
		result += fmt.Sprintf("%s`@%s`",
			s[last:indices[0]],
			s[indices[0]+1:indices[1]])
		last = indices[1]
		final = s[indices[1]:]

	}
	result += final
	return result
}

func authorizeTwitch() {
	log.Println("auth on twitch")
	client := twitch.NewAnonymousClient()

	client.OnConnect(func() {
		log.Println("connected")
	})
	client.OnPrivateMessage(handleChatMessage)

	client.Join(config.Chats...)
	if err := client.Connect(); err != nil {
		log.Panic(err)
	}
}

func handleChatMessage(message twitch.PrivateMessage) {
	// log.Print(message.Message)
	msgText := message.Message

	if strings.Contains(strings.ToLower(message.Message), "@"+strings.ToLower(config.UserName)) {
		authorizeTelegram()
		msgText = removeNicknames(msgText)
		sendMsg(
			fmt.Sprintf("chat: %s\n`@%s`: %s",
				message.Channel,
				message.User.DisplayName,
				msgText))
	} else {
		for _, trigger := range config.Triggers {
			if strings.Contains(strings.ToLower(message.Message), strings.ToLower(trigger)) {
				authorizeTelegram()
				if match, _ := regexp.Match(`@\w+`, []byte(message.Message)); match {
					msgText = removeNicknames(message.Message)
				}
				sendMsg(
					fmt.Sprintf("chat: %s\n`@%s`: %s",
						message.Channel,
						message.User.DisplayName,
						strings.Replace(
							msgText,
							trigger,
							fmt.Sprintf("`%s`", trigger),
							-1)))
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
