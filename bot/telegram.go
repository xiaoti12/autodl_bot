package bot

import (
	"autodl_bot/client"
	"autodl_bot/config"
	"fmt"
	"log"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	autodl      *client.AutoDLClient
	userConfig  map[int]*config.Config
	configMutex sync.RWMutex
}

func NewBot(token string, proxy *http.Client) (*Bot, error) {
	var api *tgbotapi.BotAPI
	var err error
	if proxy != nil {
		api, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, proxy)
	} else {
		api, err = tgbotapi.NewBotAPI(token)
	}
	if err != nil {
		return nil, err
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

func (b *Bot) initAutoDLClient(userID int) error {
	cfg := b.getUserConfig(userID)
	if cfg.AutoDLUser == "" || cfg.AutoDLPass == "" {
		return fmt.Errorf("请先设置AutoDL用户名和密码")
	}

	b.autodl = client.NewAutoDLClient(cfg.AutoDLUser, cfg.AutoDLPass)
	return nil
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
	var reply string

	switch msg.Command() {
	case "start", "help":
		reply = `支持的命令：
/user - 设置AutoDL用户名（手机号）
/password - 设置AutoDL密码
/gpuvalid - 查看所有GPU实例空闲情况`

	case "user":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带用户名，例如：/user 18900000000"
		} else {
			cfg.AutoDLUser = msg.CommandArguments()
			b.SetUserConfig(int(msg.From.ID), cfg)
			reply = "用户名设置成功"
		}

	case "password":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带密码，例如：/password 123456"
		} else {
			cfg.AutoDLPass = msg.CommandArguments()
			b.SetUserConfig(int(msg.From.ID), cfg)
			reply = "密码设置成功"

			b.initAutoDLClient(int(msg.From.ID))
		}

	case "gpuvalid":
		if b.autodl == nil {
			err := b.initAutoDLClient(int(msg.From.ID))
			if err != nil {
				reply = err.Error()
				break
			}
		}

		gpuStatus, err := b.autodl.GetGPUStatus()
		if err != nil {
			reply = fmt.Sprintf("获取GPU状态失败：%v", err)
		} else {
			reply = gpuStatus
		}

	default:
		reply = "未知命令，请使用 /help 查看支持的命令"
	}

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, reply)

	_, err := b.api.Send(replyMsg)
	if err != nil {
		log.Printf("error sending message: %v", err)
	}
}
