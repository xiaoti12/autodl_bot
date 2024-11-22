package bot

import (
	"autodl_bot/client"
	"autodl_bot/config"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	autodl      *client.AutoDLClient
	userConfig  map[int]*config.Config
	configMutex sync.RWMutex
}

func NewBot(token string) (*Bot, error) {
	api, error := tgbotapi.NewBotAPI(token)
	if error != nil {
		return nil, error
	}
	return &Bot{
		api:        api,
		userConfig: make(map[int]*config.Config),
	}, nil
}
func (b *Bot) getUserConfig(userId int) *config.Config {
	b.configMutex.RLock()
	defer b.configMutex.RUnlock()

	cfg, exist := b.userConfig[userId]
	if !exist {
		cfg = &config.Config{}
		b.userConfig[userId] = cfg
	}
	return cfg
}
func (b *Bot) SetUserConfig(userId int, cfg *config.Config) {
	b.configMutex.Lock()
	defer b.configMutex.Unlock()
	b.userConfig[userId] = cfg
}

func (b *Bot) Start() error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updatesCh := b.api.GetUpdatesChan(updateConfig)

	for update := range updatesCh {
		if update.Message == nil {
			continue
		}

		userID := update.Message.From.ID
		cfg := b.getUserConfig(int(userID))

		// process command
		if update.Message.IsCommand() {
			b.Command(update.Message, cfg)
		} else {
			// not supported command
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "未知命令，请使用 /help 查看支持的命令")
			b.api.Send(msg)
		}
	}
	return nil
}
func (b *Bot) Command(msg *tgbotapi.Message, cfg *config.Config) {

}
