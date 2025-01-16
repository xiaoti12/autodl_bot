package bot

import (
	"autodl_bot/client"
	"autodl_bot/models"
	"autodl_bot/storage"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	autodl      *client.AutoDLClient
	userConfig  map[int]*models.AutoDLConfig
	configMutex sync.RWMutex
	storage     *storage.UserStorage
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

	userStg, err := storage.NewUserStorage()
	if err != nil {
		return nil, err
	}

	userConfig, err := userStg.LoadUser()
	if err != nil {
		return nil, err
	}

	commands := []tgbotapi.BotCommand{
		{
			Command:     "user",
			Description: "设置用户名",
		},
		{
			Command:     "password",
			Description: "设置密码",
		},
		{
			Command:     "gpuvalid",
			Description: "查看GPU空闲情况",
		},
		{
			Command:     "start",
			Description: "启动GPU实例",
		},
		{
			Command:     "startcpu",
			Description: "启动GPU实例(无卡模式)",
		},
		{
			Command:     "stop",
			Description: "关闭GPU实例",
		},
		{
			Command:     "refresh",
			Description: "刷新GPU实例释放时长",
		},
		{
			Command:     "getuser",
			Description: "列出当前用户",
		},
		{
			Command:     "balance",
			Description: "查看用户余额",
		},
	}

	// 设置命令菜单
	cmdConfig := tgbotapi.NewSetMyCommands(commands...)
	_, err = api.Request(cmdConfig)
	if err != nil {
		return nil, fmt.Errorf("设置命令菜单失败: %v", err)
	}

	return &Bot{
		api:        api,
		userConfig: userConfig,
		storage:    userStg,
	}, nil
}
func (b *Bot) getUserConfig(userId int) *models.AutoDLConfig {
	b.configMutex.RLock()
	defer b.configMutex.RUnlock()

	cfg, exist := b.userConfig[userId]
	if !exist {
		cfg = &models.AutoDLConfig{}
		b.userConfig[userId] = cfg
	}
	return cfg
}
func (b *Bot) SetUserConfig(userId int, cfg *models.AutoDLConfig) {
	b.configMutex.Lock()
	defer b.configMutex.Unlock()
	b.userConfig[userId] = cfg
}

func (b *Bot) SaveAllUserConfig() error {
	b.configMutex.Lock()
	defer b.configMutex.Unlock()
	for id, cfg := range b.userConfig {
		err := b.storage.SaveUser(id, cfg.Username, cfg.Password)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) CurrentUser(userId int) string {
	cfg := b.getUserConfig(userId)
	if cfg.Username != "" {
		return "当前已设置用户: " + cfg.Username
	} else {
		return "当前未设置用户"
	}
}

func (b *Bot) initAutoDLClient(userID int) error {
	cfg := b.getUserConfig(userID)
	if cfg.Username == "" || cfg.Password == "" {
		return fmt.Errorf("请先设置AutoDL用户名和密码")
	}

	b.autodl = client.NewAutoDLClient(cfg.Username, cfg.Password)
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
func (b *Bot) Command(msg *tgbotapi.Message, cfg *models.AutoDLConfig) {
	var reply string

	switch msg.Command() {
	case "help":
		reply = `支持的命令：
/user - 设置AutoDL用户名（手机号）
/password - 设置AutoDL密码
/gpuvalid - 查看所有GPU实例空闲情况
/start - 打开实例
/startcpu - 打开实例(无卡模式)
/stop - 关闭实例
/refresh - 刷新实例释放时长
/getuser - 列出当前已设置的用户
/balance - 查看用户余额`

	case "user":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带用户名，例如：/user 18900000000"
		} else {
			cfg.Username = msg.CommandArguments()
			b.SetUserConfig(int(msg.From.ID), cfg)
			reply = "用户名设置成功"
		}

	case "password":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带密码，例如：/password 123456"
		} else {
			rawPassword := msg.CommandArguments()
			cfg.Password = client.HashPassword(rawPassword)
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

	case "start", "startcpu":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带实例UUID，例如：/start xx-yy"
		} else if b.autodl == nil {
			err := b.initAutoDLClient(int(msg.From.ID))
			if err != nil {
				reply = err.Error()
				break
			}
		} else {
			useCPU := msg.Command() == "startcpu"
			uuid := msg.CommandArguments()
			err := b.autodl.PowerOn(uuid, useCPU)
			if err != nil {
				reply = err.Error()
			} else {
				reply = fmt.Sprintf("实例 %s 开机成功", uuid)
			}
		}
	case "stop":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带实例UUID，例如：/stop xx-yy"
		} else if b.autodl == nil {
			err := b.initAutoDLClient(int(msg.From.ID))
			if err != nil {
				reply = err.Error()
				break
			}
		} else {
			uuid := msg.CommandArguments()
			err := b.autodl.PowerOff(uuid)
			if err != nil {
				reply = err.Error()
			} else {
				reply = fmt.Sprintf("实例 %s 关机成功", uuid)
			}
		}
	case "refresh":
		if msg.CommandArguments() == "" {
			reply = "请在命令后附带实例UUID，例如：/refresh xx-yy"
		} else if b.autodl == nil {
			err := b.initAutoDLClient(int(msg.From.ID))
			if err != nil {
				reply = err.Error()
				break
			}
		} else {
			uuid := msg.CommandArguments()
			err := b.autodl.PowerOn(uuid, true)
			if err != nil {
				reply = err.Error()
			} else {
				reply = fmt.Sprintf("实例 %s 无卡模式开机成功，10秒后关机", uuid)
				go func() {
					// 10秒后关机
					<-time.After(10 * time.Second)
					err := b.autodl.PowerOff(uuid)
					if err != nil {
						log.Printf("刷新实例 %s 释放时长失败: %v", uuid, err)
					}
				}()
			}
		}
	case "balance":
		if b.autodl == nil {
			err := b.initAutoDLClient(int(msg.From.ID))
			if err != nil {
				reply = err.Error()
				break
			}
		}
		balance, err := b.autodl.GetBalance()
		if err != nil {
			reply = err.Error()
		} else {
			reply = fmt.Sprintf("当前余额: %.2f元", balance)
		}

	case "getuser":
		reply = b.CurrentUser(int(msg.From.ID))

	default:
		reply = "未知命令，请使用 /help 查看支持的命令"
	}

	replyMsg := tgbotapi.NewMessage(msg.Chat.ID, reply)

	_, err := b.api.Send(replyMsg)
	if err != nil {
		log.Printf("error sending message: %v", err)
	}
}
