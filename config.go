package wayback

type IPFS struct {
	Host   string
	Port   uint
	UseTor bool
}

type Config struct {
	Token  string
	ChatID string
	Debug  bool
	IPFS   *IPFS
}

func NewConfig(token string, debug bool, chatid string, ipfs *IPFS) *Config {
	conf := &Config{
		Token:  token,
		ChatID: chatid,
		Debug:  debug,
		IPFS:   ipfs,
	}

	return conf
}
