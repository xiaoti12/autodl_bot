package config

type Config struct {
	TelegramToken string
	AutoDLUser    string
	AutoDLPass    string
}

var GlobalConfig Config
