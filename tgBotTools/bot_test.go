package tgBotTools

import (
	"fmt"
	"testing"
)

type test struct {
	*testing.T
}

func (t *test) Command(b *BotApi, m *Msg) any {
	//fmt.Printf("命令 %+v", m.User)
	println(fmt.Sprintf("命令 %+v", m.User))
	println(m.Text)
	return "ok"
}
func (t *test) Message(b *BotApi, m *Msg) any {

	println(fmt.Sprintf("信息 %+v", m.User))
	println(m.Text)
	return "https://fanyi.baidu.com/#en/zh/?AAGWbfmyuLfca21Mnnvml4H4GW8Mi8CfKYgAAGWbfmyuLfca21Mnnvml4H4GW8Mi8CfKYgAAGWbfmyuLfca21Mnnvml4H4GW8Mi8CfKYgAAGWbfmyuLfca21Mnnvml4H4GW8Mi8CfKYg"
}
func TestNewBot(t *testing.T) {

	bot, err := NewBotApp("5426849660:AAGWbfmyuLfca21Mnnvml4H4GW8Mi8CfKYg", true)
	if err != nil {
		panic(err)
	}

	//bot.SetCommands([]tgbotapi.BotCommand{{"/test", "测试"}, {"/user", "账户信息"}}...)
	bot.SendMsg(2073804804, "测试啊")
	bot.Run(30, new(test))

}
