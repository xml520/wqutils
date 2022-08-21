package tgBotTools

import (
	"errors"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type User struct {
	TelegramID int64  // 电报id
	Nickname   string // 昵称
	Username   string // 用户名
}
type Msg struct {
	ID   int
	Text string
	*User
}
type Handle interface {
	Command(*BotApi, *Msg) any
	Message(*BotApi, *Msg) any
}
type BotApi struct {
	api *tgbotapi.BotAPI
}

func (b *BotApi) Api() *tgbotapi.BotAPI {
	return b.api
}

// Run 监听信息
func (b *BotApi) Run(timeout int, handle Handle) (err error) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
			err = errors.New("机器人程序奔溃")
		}
	}()
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = timeout
	updates := b.Api().GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		go func(msg tgbotapi.Update) {
			var res any
			u := &User{Nickname: msg.FromChat().FirstName, TelegramID: msg.FromChat().ID, Username: msg.FromChat().UserName}
			if msg.Message.IsCommand() {
				res = handle.Command(b, &Msg{ID: msg.Message.MessageID, Text: msg.Message.Text, User: u})
			} else {
				res = handle.Message(b, &Msg{ID: msg.Message.MessageID, Text: msg.Message.Text, User: u})
			}
			if res != nil {
				var m tgbotapi.MessageConfig
				switch res.(type) {
				case string:
					m = tgbotapi.NewMessage(update.Message.Chat.ID, res.(string))
				default:
					m = tgbotapi.NewMessage(update.Message.Chat.ID, "系统内部错误")
				}
				if _, _err := b.Api().Send(m); _err != nil {
					fmt.Println("信息发送失败", res)
				}
			}
		}(update)
	}
	return nil
}

// SetCommands 设置命令
func (b *BotApi) SetCommands(s ...tgbotapi.BotCommand) error {
	menu := tgbotapi.SetMyCommandsConfig{
		Commands: s,
		Scope:    nil,
	}
	_, err := b.api.Request(menu)
	return err
}

// SendMsg 发送信息
func (b *BotApi) SendMsg(uid int64, text string) (err error) {
	_, err = b.api.Send(tgbotapi.NewMessage(uid, text))
	return
}

// User 获取机器人信息
func (b *BotApi) User() *tgbotapi.User {
	return &b.api.Self
}
func NewBotApp(apiToken string, debug bool) (*BotApi, error) {
	var api *tgbotapi.BotAPI
	var err error
	if debug {
		api, err = tgbotapi.NewBotAPIWithAPIEndpoint(apiToken, "https://test.redapple0114.com/bot%s/%s")
	} else {
		api, err = tgbotapi.NewBotAPI(apiToken)
	}
	if err != nil {
		return nil, err
	}
	//api.Debug = debug
	return &BotApi{api: api}, err
}
