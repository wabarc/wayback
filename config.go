package wayback

type IPFSRV struct {
	Host   string
	Port   uint
	Mode   string
	UseTor bool
}

type Config struct {
	Token   string
	ChatID  string
	Debug   bool
	IPFS    *IPFSRV
	handler map[string]bool
}

func NewConfig(token string, debug bool, chatid string, h map[string]bool, ipfs *IPFSRV) *Config {
	conf := &Config{
		Token:   token,
		ChatID:  chatid,
		Debug:   debug,
		IPFS:    ipfs,
		handler: h,
	}

	return conf
}
