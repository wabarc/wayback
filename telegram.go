package wayback

import (
	"bytes"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/wabarc/archive.is/pkg"
	"github.com/wabarc/archive.org/pkg"
	"log"
	"regexp"
	"text/template"
)

type Message struct {
	Archiver []string
	URL      [][]string
}

const tmpl = `{{range $i, $name := .Archiver}}{{$name}}:
{{ range $url := index $.URL $i -}}
* {{ $url }}
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
		iaURL := ia.Wayback(url)
		isURL := is.Wayback(url)

		if len(iaURL) > 0 || len(isURL) > 0 {
			vars := &Message{
				Archiver: []string{"Internet Archive", "archive.today"},
				URL:      [][]string{iaURL, isURL},
			}
			replyText := message(vars)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
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
