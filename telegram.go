package wayback

import (
	"bytes"
	"log"
	"regexp"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type collect struct {
	Arc []string
	Dst []map[string]string
}

func (cfg *Config) Telegram() {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = cfg.Debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		text := update.Message.Text
		uri := []string{}
		re := regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
		match := re.FindAllString(text, -1)
		for _, el := range match {
			uri = append(uri, el)
		}
		if len(uri) == 0 {
			continue
		}

		c := &collect{}
		h := Handle{URI: uri}
		for hd, do := range cfg.handler {
			switch {
			case hd == "ia" && do:
				c.Arc = append(c.Arc, "Internet Archive")
				c.Dst = append(c.Dst, h.IA())
			case hd == "is" && do:
				c.Arc = append(c.Arc, "Archive Today")
				c.Dst = append(c.Dst, h.IS())
			case hd == "ip" && do:
				h.IPFS = cfg.IPFS
				c.Arc = append(c.Arc, "IPFS(beta)")
				c.Dst = append(c.Dst, h.WBIPFS())
			}
		}

		replyText := message(c)
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

func message(vars *collect) string {
	var tmplBytes bytes.Buffer

	const tmpl = `{{range $i, $name := .Arc}}<b>{{ $name }}</b>:
{{ range $url := index $.Dst $i -}}
• {{ $url }}
{{end}}
{{end}}`

	tpl, err := template.New("message").Parse(tmpl)
	if err != nil {
		return ""
	}

	err = tpl.Execute(&tmplBytes, vars)
	if err != nil {
		return ""
	}

	return tmplBytes.String()
}
