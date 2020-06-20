package wayback

type Config struct {
	Token  string
	ChatID string
	Debug  bool
}

func NewConfig(token string, debug bool, chatid string) *Config {
	conf := &Config{
		Token:  token,
		ChatID: chatid,
		Debug:  debug,
	}

	return conf
}
