package wayback

type Config struct {
	Token  string
	Debug  bool
}

func NewConfig(token string, debug bool) *Config {
	conf := &Config{
		Token:  token,
		Debug:  debug,
	}

	return conf
}
