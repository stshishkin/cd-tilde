// Copyright © 2022 Stepan Shishkin. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var clocks = [12]string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"}
var owner_id int64
var file_msg map[int64]int = make(map[int64]int)
var tic map[int64]int = make(map[int64]int)
var tic_msg map[int64]int = make(map[int64]int)
var value map[int64]int = make(map[int64]int)
var on map[int64]bool = make(map[int64]bool)
var off map[int64]bool = make(map[int64]bool)

func load_config() (map[string]int, error) {
	config := make(map[string]int)
	config_file, err := os.Open("config.json")
	defer config_file.Close()
	if err != nil {
		return config, err
	}
	byteValue, _ := ioutil.ReadAll(config_file)
	err = json.Unmarshal(byteValue, &config)
	return config, err
}

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

func call_on(wg *sync.WaitGroup, bot *tgbotapi.BotAPI, chat int64, t int) {
	defer wg.Done()
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

	tic[chat] = s
	file_msg[chat] = c.MessageID
	if err != nil {
		go countdown(wg, bot, chat, t)
	}
}

func countdown(wg *sync.WaitGroup, bot *tgbotapi.BotAPI, chat int64, t int) {
	defer wg.Done()

	s := tic[chat]
	msg := tgbotapi.NewMessage(chat, fmt.Sprintf("VPN will be working for 🕛 %02d:%02d", t/60, t%60))
	m, err := bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
	tic_msg[chat] = m.MessageID
	s = t - s
	for s > 0 {
		msg1 := tgbotapi.NewEditMessageText(chat, m.MessageID, fmt.Sprintf("VPN will be working for %s %02d:%02d", clocks[s%len(clocks)], s/60, s%60))

		if _, err = bot.Send(msg1); err != nil {
			fmt.Println(err)
			break
		}

		time.Sleep(1 * time.Second)
		s--
	}
	if s == 0 {
		go call_off(wg, bot, chat)
		on[chat] = false
	}
}

func call_off(wg *sync.WaitGroup, bot *tgbotapi.BotAPI, chat int64) {
	defer wg.Done()
	from := fmt.Sprintf("%d", chat)

	del := tgbotapi.NewDeleteMessage(chat, file_msg[chat])
	_, err := bot.Send(del)
	if err != nil {
		fmt.Println(err)
	}

	if tic_msg[chat] != 0 {
		del = tgbotapi.NewDeleteMessage(chat, tic_msg[chat])
		_, err = bot.Send(del)
		if err != nil {
			fmt.Println(err)
		}
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

	off[chat] = false
}

func start(wg *sync.WaitGroup, bot *tgbotapi.BotAPI, chat int64, time int, allow bool, request string) {
	defer wg.Done()
	var text string
	if allow {
		text = `This bot creates a temporary Russian standalone VPN server for you for %02d:%02d minutes and gives a .ovpn file for OpenVPN compatible clients.
Use it with caution, cause it's in Russian jurisdiction. 

I know these commands:

/start /help - shows this message
/on - creates VPN server. Please note that creating a real server may take 1-2 minutes.
/off - destroys your VPN server. It will be destroyed after timeout anyway.

Welcome back home, son.`
		text = fmt.Sprintf(text, time/60, time%60)
	} else {
		text = `Hi stranger. This is a private bot.
The owner will allow your access if it's necessary.`

		msg := tgbotapi.NewMessage(int64(100166704), request)
		_, err := bot.Send(msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	msg := tgbotapi.NewMessage(chat, text)
	_, err := bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	token := os.Getenv("BOT_APITOKEN")
	domain := os.Getenv("BOT_DOMAIN")
	port := os.Getenv("BOT_PORT")
	cert := os.Getenv("BOT_CERT")
	key := os.Getenv("BOT_KEY")
	owner := os.Getenv("BOT_OWNER")

	var wg sync.WaitGroup
	var err error

	if token == "" || domain == "" || cert == "" || key == "" || owner == "" {
		fmt.Println("Missing startup environment variable. Please note, you have to set up BOT_APITOKEN, BOT_DOMAIN, BOT_CERT, BOT_KEY and BOT_OWNER. ")
		os.Exit(1)
	}

	if port == "" {
		port = "8443"
	}

	if _, err = os.Stat(cert); err != nil {
		fmt.Println("Missing file in BOT_CERT environment variable.")
		os.Exit(1)
	}

	if _, err = os.Stat(key); err != nil {
		fmt.Println("Missing file in BOT_KEY environment variable.")
		os.Exit(1)
	}

	if _, err = os.Stat("config.json"); err != nil {
		fmt.Println("Missing config.json file in bot folder.")
		os.Exit(1)
	}

	if owner_id, err = strconv.ParseInt(owner, 10, 64); err != nil {
		fmt.Println("BOT_OWNER variable should contain a valid id of telegram user")
		os.Exit(1)
	}

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		fmt.Println(err)
	}

	// bot.Debug = true

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhook("https://" + domain + ":" + port + "/" + token)

	_, err = bot.Request(wh)
	if err != nil {
		fmt.Println(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		fmt.Println(err)
	}

	if info.LastErrorDate != 0 {
		fmt.Printf("Telegram callback failed: %s\n", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + token)
	go http.ListenAndServeTLS("0.0.0.0:"+port, cert, key, nil)

	for update := range updates {
		config, err := load_config()
		if err != nil {
			fmt.Println(err)
		}
		wg.Add(len(config))
		if temp, err := config[fmt.Sprintf("%d", update.Message.Chat.ID)]; err {
			value[update.Message.Chat.ID] = temp
			fmt.Printf("%+v\n", update.Message)
			if value[update.Message.Chat.ID] < 0 {
				value[update.Message.Chat.ID] = 600
			}
			if update.Message.Text == "/start" || update.Message.Text == "/help" {
				start(&wg, bot, update.Message.Chat.ID, value[update.Message.Chat.ID], true, "")
			} else if update.Message.Text == "/on" {
				if on[update.Message.Chat.ID] {
					continue
				}
				on[update.Message.Chat.ID] = true
				go call_on(&wg, bot, update.Message.Chat.ID, value[update.Message.Chat.ID])

			} else if update.Message.Text == "/off" {
				if off[update.Message.Chat.ID] || !on[update.Message.Chat.ID] {
					continue
				}
				off[update.Message.Chat.ID] = true
				go call_off(&wg, bot, update.Message.Chat.ID)

			} else {
				fmt.Printf("wrong: %s\n", update.Message.Text)
				fmt.Printf("%+v\n", update.Message.From.ID)
			}
		} else {
			fmt.Printf("%+v\n", config[fmt.Sprintf("%d", update.Message.Chat.ID)])
			fmt.Printf("%d\n", update.Message.Chat.ID)
			fmt.Println(err)
			if update.Message.Text == "/start" {
				start(&wg, bot, update.Message.Chat.ID, value[update.Message.Chat.ID], false, fmt.Sprintf("%s(@%s) %d", update.Message.Chat.FirstName, update.Message.Chat.UserName, update.Message.Chat.ID))
			}
		}
	}
	wg.Wait()
}
