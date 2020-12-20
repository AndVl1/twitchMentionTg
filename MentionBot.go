package main

import (
	"encoding/json"
	"fmt"
	"github.com/gempir/go-twitch-irc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
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
	log.Println(msg)
	authorizeTelegram()
	message := tgbotapi.NewMessage(config.ChatId, msg)
	message.ParseMode = "HTML"
	if _, err := botApi.Send(message); err != nil {
		log.Panic(err)
	}
}

func authorize() {
	configText, err := ioutil.ReadFile("config.json")
	if err == nil {
		if err := json.Unmarshal(configText, &config); err != nil {
			log.Fatal(err)
		}
	} else {
		getFromEnvVariables(&config)
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
	final := s
	var matches = re.FindAllStringSubmatchIndex(s, -1)
	last := 0
	for _, indices := range matches {
		result += fmt.Sprintf("%s<code>@%s</code>",
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
	msgText := message.Message

	if strings.Contains(strings.ToLower(message.Message), "@"+strings.ToLower(config.UserName)) {
		msgText = removeNicknames(msgText)
		log.Println(msgText)
		sendMsg(
			fmt.Sprintf("chat: %s\n<code>@%s</code>: %s",
				message.Channel,
				message.User.DisplayName,
				msgText))
	} else {
		for _, trigger := range config.Triggers {
			if strings.Contains(strings.ToLower(message.Message), strings.ToLower(trigger)) {
				if match, _ := regexp.Match(`@\w+`, []byte(message.Message)); match {
					msgText = removeNicknames(message.Message)
				}
				sendMsg(
					fmt.Sprintf("chat: %s\n<code>@%s</code>: %s",
						message.Channel,
						message.User.DisplayName,
						strings.Replace(
							msgText,
							trigger,
							fmt.Sprintf("<code>%s</code>", trigger),
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

func getFromEnvVariables(config *Cfg) {
	config.UserName = os.Getenv("user_name")
	config.ChatId, _ = strconv.ParseInt(os.Getenv("chat_id"), 10, 64)
	config.ApiKey = os.Getenv("")
	config.Triggers = strings.Split(os.Getenv("triggers"), ",")
	config.Chats = strings.Split(os.Getenv("chats"), ",")
}
