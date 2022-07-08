package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var clocks = [12]string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"}
var s int

func cmd_handler(bot *tgbotapi.BotAPI, chat int64, msg int, cmd string) int {

	s := 1

	fmt.Println(cmd)
	cmd_array := strings.Split(cmd, " ")

	cmd_exec := exec.Command(cmd_array[0], cmd_array[1:]...)
	cmd_exec.Dir = ".."

	err := cmd_exec.Start()
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		err = cmd_exec.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}()

	for {
		process, err := os.FindProcess(cmd_exec.Process.Pid)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = process.Signal(syscall.Signal(0))
		if err != nil && err.Error() == "os: process already finished" {
			fmt.Printf("process.Signal on pid %d returned: '%v'\n", cmd_exec.Process.Pid, err)
			fmt.Printf("%d\n", msg)
			break
		} else {
			msg1 := tgbotapi.NewEditMessageText(chat, msg, fmt.Sprintf("please, wait while kitty rolling. %s %d roll", clocks[s%len(clocks)], s))

			if _, err = bot.Send(msg1); err != nil {
				fmt.Println(err)
			}

			time.Sleep(1 * time.Second)
			s++
		}
	}
	return s
}

func call_on(bot *tgbotapi.BotAPI, chat int64) int {
	from := fmt.Sprintf("%d", chat)
	msg := tgbotapi.NewMessage(chat, "please, wait while kitty rolling. 🕛 1 roll")

	m, err := bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
	m_id := fmt.Sprintf("%d", m.MessageID)
	cat := tgbotapi.NewDocument(chat, tgbotapi.FilePath("kitty-roll.mp4"))

	c, err := bot.Send(cat)
	if err != nil {
		fmt.Println(err)
	}
	cat_id := fmt.Sprintf("%d", c.MessageID)

	s := cmd_handler(bot, chat, m.MessageID, "terraform apply -var=telegram-chat="+from+" -var=countdown-msg="+m_id+" -var=file-msg="+cat_id+" -input=false -auto-approve -state="+from+".tfstate")

	for err == nil {
		msg1 := tgbotapi.NewEditMessageText(chat, m.MessageID, fmt.Sprintf("please, wait while kitty rolling. %s %d roll", clocks[s%len(clocks)], s))

		if _, err = bot.Send(msg1); err != nil {
			fmt.Println(err)
			break
		}

		time.Sleep(1 * time.Second)
		s++
	}
	return c.MessageID
}
func call_off(bot *tgbotapi.BotAPI, chat int64, file_msg int) {
	from := fmt.Sprintf("%d", chat)
	del := tgbotapi.NewDeleteMessage(chat, file_msg)
	_, err := bot.Send(del)
	if err != nil {
		fmt.Println(err)
	}
	msg := tgbotapi.NewMessage(chat, "please, wait while kitty rolling. 🕛 1 roll")

	m, err := bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
	cat := tgbotapi.NewDocument(chat, tgbotapi.FilePath("kitty-unroll.mp4"))

	c, err := bot.Send(cat)
	if err != nil {
		fmt.Println(err)
	}
	cmd_handler(bot, chat, m.MessageID, "terraform destroy -var=telegram-chat="+from+" -var=countdown-msg=0 -var=file-msg=0 -input=false -auto-approve -state="+from+".tfstate")

	del = tgbotapi.NewDeleteMessage(chat, m.MessageID)
	_, err = bot.Send(del)
	if err != nil {
		fmt.Println(err)
	}
	del = tgbotapi.NewDeleteMessage(chat, c.MessageID)
	_, err = bot.Send(del)
	if err != nil {
		fmt.Println(err)
	}
}
func main() {
	bot, err := tgbotapi.NewBotAPI("YOUR_SECRET_TOKEN")
	if err != nil {
		fmt.Println(err)
	}

	bot.Debug = true

	fmt.Println("Authorized on account %s", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhook("https://your-public-domain:8443/" + bot.Token)

	_, err = bot.Request(wh)
	if err != nil {
		fmt.Println(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		fmt.Println(err)
	}

	if info.LastErrorDate != 0 {
		fmt.Println("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "fullchain.pem", "privkey.pem", nil)
	var on, off int64
	var file_msg int
	on = 0
	off = 0
	file_msg = 0
	for update := range updates {
		fmt.Println("%+v\n", update.Message)
		if update.Message.Text == "/on" {
			if on == update.Message.Chat.ID {
				continue
			}
			on = update.Message.Chat.ID
			file_msg = call_on(bot, update.Message.Chat.ID)
		} else if update.Message.Text == "/off" {
			if on != update.Message.Chat.ID || off == update.Message.Chat.ID {
				continue
			}
			off = update.Message.Chat.ID
			call_off(bot, update.Message.Chat.ID, file_msg)
		} else {
			fmt.Println("wrong: %s", update.Message.Text)
			fmt.Println("%+v", update.Message.From.ID)
		}
	}
}
