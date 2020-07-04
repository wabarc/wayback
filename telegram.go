package wayback

import (
	"bytes"
	"log"
	"regexp"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wabarc/archive.is/pkg"
	"github.com/wabarc/archive.org/pkg"
	"github.com/wabarc/wbipfs"
)

type Message struct {
	Archiver []string
	URL      []map[string]string
}

const tmpl = `{{range $i, $name := .Archiver}}<b>{{ $name }}</b>:
{{ range $url := index $.URL $i -}}
• {{ $url }}
{{end}}
{{end}}`

func (cfg *Config) Telegram() {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = cfg.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	iaWbrc := &ia.Archiver{}
	isWbrc := &is.Archiver{}
	ipfsWbrc := &wbipfs.Archiver{IPFSHost: cfg.IPFS.Host, IPFSPort: cfg.IPFS.Port, UseTor: cfg.IPFS.UseTor}
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		text := update.Message.Text
		url := []string{}
		re := regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
		match := re.FindAllString(text, -1)
		for _, el := range match {
			url = append(url, el)
		}
		iaURL, _ := iaWbrc.Wayback(url)
		isURL, _ := isWbrc.Wayback(url)
		ipfsURL, _ := ipfsWbrc.Wayback(url)

		if len(iaURL) > 0 || len(isURL) > 0 {
			vars := &Message{
				Archiver: []string{"Internet Archive", "archive.today", "IPFS(alpha)"},
				URL:      []map[string]string{iaURL, isURL, ipfsURL},
			}
			replyText := message(vars)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.ParseMode = "html"

			bot.Send(msg)

			if len(cfg.ChatID) > 0 {
				msg = tgbotapi.NewMessageToChannel("@"+cfg.ChatID, replyText)
				msg.ParseMode = "html"
				bot.Send(msg)
			}
		}
	}
}

func message(vars *Message) string {
	var tmplBytes bytes.Buffer

	tmpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		return ""
	}

	err = tmpl.Execute(&tmplBytes, vars)
	if err != nil {
		return ""
	}

	return tmplBytes.String()
}
