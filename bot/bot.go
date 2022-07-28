package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var clocks = [12]string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"}
var file_msg map[int64]int = make(map[int64]int)
var tic map[int64]int = make(map[int64]int)
var tic_msg map[int64]int = make(map[int64]int)
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
	msg := tgbotapi.NewMessage(chat, "VPN will working for 🕛")
	m, err := bot.Send(msg)
	if err != nil {
		fmt.Println(err)
	}
	tic_msg[chat] = m.MessageID
	s = t - s
	for s > 0 {
		msg1 := tgbotapi.NewEditMessageText(chat, m.MessageID, fmt.Sprintf("VPN will working for %s %02d:%02d", clocks[s%len(clocks)], s/60, s%60))

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

func main() {
	token := os.Getenv("BOT_APITOKEN")
	domain := os.Getenv("BOT_DOMAIN")
	port := os.Getenv("BOT_PORT")
	cert := os.Getenv("BOT_CERT")
	key := os.Getenv("BOT_KEY")

	var wg sync.WaitGroup

	if token == "" || domain == "" || cert == "" || key == "" {
		fmt.Println("Missing startup environment variable. Please note, you have to set up BOT_APITOKEN, BOT_DOMAIN, BOT_CERT and BOT_KEY. ")
		os.Exit(1)
	}

	if port == "" {
		port = "8443"
	}

	bot, err := tgbotapi.NewBotAPI(token)
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
		if time, err := config[fmt.Sprintf("%d", update.Message.Chat.ID)]; err {
			fmt.Printf("%+v\n", update.Message)
			if update.Message.Text == "/on" {
				if on[update.Message.Chat.ID] {
					continue
				}
				on[update.Message.Chat.ID] = true
				if time < 0 {
					time = 300
				}
				go call_on(&wg, bot, update.Message.Chat.ID, time)

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
		}
	}
	wg.Wait()
}
