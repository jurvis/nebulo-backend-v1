package config

import (
	"code.google.com/p/gcfg"
)

type DbCfg struct {
	Database struct {
		Username string
		Password string
		Dbname   string
		Host     string
	}
}

type EmailCfg struct {
	Alert struct {
		Address string
		Password string
		SMTPServer string
		SMTPPort string
		Recipients string
	}
}

type PushCfg struct {
	GCM struct {
		ApiKey string
	}
}

type NewRelicCfg struct {
	License struct {
		Key string
	}
}

type TwCfg struct {
	Application struct {
		ApiKey    string
		ApiSecret string
	}
	Consumer struct {
		Token  string
		Secret string
	}
}

// Read .gcfg file into struct
func DbConfig() DbCfg {
	var cfg DbCfg
	gcfg.ReadFileInto(&cfg, "dbconfig.gcfg")
	return cfg
}

func PushConfig() PushCfg {
	var cfg PushCfg
	gcfg.ReadFileInto(&cfg, "pushconfig.gcfg")
	return cfg
}

func EmailConfig() EmailCfg {
	var cfg EmailCfg
	gcfg.ReadFileInto(&cfg, "emailconfig.gcfg")
	return cfg
}

func NewRelicConfig() NewRelicCfg {
	var cfg NewRelicCfg
	gcfg.ReadFileInto(&cfg, "newrelic.gcfg")
	return cfg
}

func TwitterConfig() TwCfg {
	var cfg TwCfg
	gcfg.ReadFileInto(&cfg, "twitterconfig.gcfg")
	return cfg
}