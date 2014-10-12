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

func TwitterConfig() TwCfg {
	var cfg TwCfg
	gcfg.ReadFileInto(&cfg, "twitterconfig.gcfg")
	return cfg
}
