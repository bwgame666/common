package libs

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"net/http"
	"net/url"
)

func TelegramNotice(botId string, chatId int64, msg string, proxy string) {
	httpClient := &http.Client{}
	if len(proxy) > 0 && proxy != "0.0.0.0" {
		proxyUrl, _ := url.Parse("http://" + proxy)
		httpClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
		fmt.Println("telegram use proxy")
	}
	bot, err := tgbotapi.NewBotAPIWithClient(botId, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		fmt.Printf("bot init error: %s\n", err.Error())
	}

	bot.Debug = false
	content := tgbotapi.NewMessage(chatId, "")
	content.Text = msg
	if _, err := bot.Send(content); err != nil {
		fmt.Println("telegram send error : ", err.Error())
	}
}
