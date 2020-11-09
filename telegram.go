package wayback

import (
	"bytes"
	"context"
	"log"
	"regexp"
	"text/template"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/sync/errgroup"
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

	g, _ := errgroup.WithContext(context.Background())
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		update := update
		text := update.Message.Text
		urls := []string{}
		re := regexp.MustCompile(`https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,4}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
		match := re.FindAllString(text, -1)
		for _, el := range match {
			urls = append(urls, el)
		}
		if len(urls) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "URL no found.")
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
			continue
		}

		g.Go(func() error {
			msgID, archives, err := cfg.archive(update.Message.MessageID, urls)
			if err != nil {
				return err
			}

			replyText := message(archives)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
			msg.ReplyToMessageID = msgID
			msg.ParseMode = "html"

			bot.Send(msg)

			if len(cfg.ChatID) > 0 {
				msg = tgbotapi.NewMessageToChannel("@"+cfg.ChatID, replyText)
				msg.ParseMode = "html"
				bot.Send(msg)
			}
			return nil
		})
	}
}

func (cfg *Config) archive(msgid int, urls []string) (int, *collect, error) {
	c := &collect{}

	h := Handle{URI: urls}
	for hd, do := range cfg.handler {
		switch {
		case hd == "ia" && do:
			c.Arc = append(c.Arc, "<a href='https://web.archive.org/'>Internet Archive</a>")
			c.Dst = append(c.Dst, h.IA())
		case hd == "is" && do:
			c.Arc = append(c.Arc, "<a href='https://archive.today/'>Archive Today</a>")
			c.Dst = append(c.Dst, h.IS())
		case hd == "ip" && do:
			h.IPFS = cfg.IPFS
			c.Arc = append(c.Arc, "<a href='https://ipfs.github.io/public-gateway-checker/'>IPFS</a>")
			c.Dst = append(c.Dst, h.WBIPFS())
		}
	}

	return msgid, c, nil
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
