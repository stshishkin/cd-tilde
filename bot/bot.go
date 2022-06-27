package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	bt "github.com/SakoDroid/telego"
	cfg "github.com/SakoDroid/telego/configs"
)

var bot *bt.Bot

func main() {
	cf, _ := cfg.Load()

	var err error

	bot, err = bt.NewBot(cf)

	if err == nil {

		err = bot.Run()

		if err == nil {
			start()
		} else {
			fmt.Println(err)
		}
	} else {
		fmt.Println(err)
	}

}

func start() {
	clocks := [12]string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"}

	messageChannel, _ := bot.AdvancedMode().RegisterChannel("", "message")

	kb := bot.CreateKeyboard(false, false, false, "")

	kb.AddButton("On", 1)
	for {
		up := <-*messageChannel
		from := fmt.Sprintf("%d", up.Message.From.Id)

		if up.Message.Text == "/on" {

			kb := bot.CreateInlineKeyboard()
			kb.AddCallbackButton(clocks[0], "callback data 1", 1)
			msg, err := bot.AdvancedMode().ASendMessage(up.Message.Chat.Id, "00:00", "", 0, false, false, nil, false, false, kb)
			if err != nil {
				fmt.Println(err)
			}

			editor := bot.GetMsgEditor(up.Message.Chat.Id)
			s := 1

			cmd := exec.Command("terraform", "apply", "-var=telegram-chat="+from, "-input=false", "-auto-approve", "-state="+from+".tfstate")
			cmd.Dir = ".."

			err = cmd.Start()
			if err != nil {
				fmt.Println(err)
			}

			go func() {
				err = cmd.Wait()
				if err != nil {
					fmt.Println(err)
				}
			}()

			for {
				process, err := os.FindProcess(cmd.Process.Pid)
				if err != nil {
					fmt.Printf("Failed to find process: %s\n", err)
					break
				}

				err = process.Signal(syscall.Signal(0))
				if err != nil && err.Error() == "os: process already finished" {
					fmt.Printf("process.Signal on pid %d returned: '%v'\n", cmd.Process.Pid, err)
					break
				} else {
					kb = bot.CreateInlineKeyboard()
					kb.AddCallbackButton(clocks[s%len(clocks)], "callback data 1", 1)
					_, err1 := editor.EditText(msg.Result.MessageId, fmt.Sprintf("%02d:%02d", s/60, s%60), "", "", nil, false, kb)
					if err1 != nil {
						fmt.Println(err1)
					}
					time.Sleep(1 * time.Second)
					s++
				}
			}
			editor.DeleteMessage(msg.Result.MessageId)
		} else {
			if up.Message.Text == "/off" {

				// TODO: add delete sended file
				kb := bot.CreateInlineKeyboard()
				kb.AddCallbackButton(clocks[0], "callback data 1", 1)
				msg, err := bot.AdvancedMode().ASendMessage(up.Message.Chat.Id, "00:00", "", 0, false, false, nil, false, false, kb)
				if err != nil {
					fmt.Println(err)
				}

				editor := bot.GetMsgEditor(up.Message.Chat.Id)
				s := 1

				cmd := exec.Command("terraform", "destroy", "-var=telegram-chat="+from, "-input=false", "-auto-approve", "-state="+from+".tfstate") //terraform apply -input=false -auto-approve
				cmd.Dir = ".."

				err = cmd.Start()
				if err != nil {
					fmt.Println(err)
				}

				go func() {
					err = cmd.Wait()
					if err != nil {
						fmt.Println(err)
					}
				}()

				for {
					process, err := os.FindProcess(cmd.Process.Pid)
					if err != nil {
						fmt.Printf("Failed to find process: %s\n", err)
						break
					}

					err = process.Signal(syscall.Signal(0))
					if err != nil && err.Error() == "os: process already finished" {
						fmt.Printf("process.Signal on pid %d returned: '%v'\n", cmd.Process.Pid, err)
						break
					} else {
						kb = bot.CreateInlineKeyboard()
						kb.AddCallbackButton(clocks[s%len(clocks)], "callback data 1", 1)
						_, err1 := editor.EditText(msg.Result.MessageId, fmt.Sprintf("%02d:%02d", s/60, s%60), "", "", nil, false, kb)
						if err1 != nil {
							fmt.Println(err1)
						}
						time.Sleep(1 * time.Second)
						s++
					}
				}
				editor.DeleteMessage(msg.Result.MessageId)
			}
		}

		fmt.Println(up.Message.Text)
	}
}
