package main

import (
	"autodl_bot/bot"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	tokenFlag = flag.String("token", "", "telegram bot token")
)

func setupLogger() (*os.File, error) {
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return nil, fmt.Errorf("无法创建日志目录: %v", err)
	}

	logFile, err := os.OpenFile(
		"logs/bot.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		os.ModePerm,
	)
	if err != nil {
		return nil, fmt.Errorf("无法创建日志文件: %v", err)
	}

	// 同时输出到标准输出和日志文件
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile, nil
}
func getToken() string {
	if *tokenFlag != "" {
		return *tokenFlag
	}

	envToken := os.Getenv("BOT_TOKEN")
	if envToken != "" {
		return envToken
	}

	log.Fatal("无法获取telegram bot token，请使用--token参数或设置BOT_TOKEN环境变量")
	return ""
}

func getProxyClient() *http.Client {
	proxyURL, err := url.Parse("http://127.0.0.1:7890")
	if err != nil {
		log.Fatal(err)
		return nil
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	return client
}

func main() {
	flag.Parse()

	logger, err := setupLogger()
	if err != nil {
		log.Fatalf("无法设置日志: %v", err)
	}
	defer logger.Close()

	tgToken := getToken()
	tgbot, err := bot.NewBot(tgToken, getProxyClient())
	if err != nil {
		log.Fatalf("无法创建telegram bot: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Bot已启动，时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))

	errCh := make(chan error, 1)
	go func() {
		errCh <- tgbot.Start()
	}()

	select {
	case sig := <-sigCh:
		log.Printf("接收到退出信号：%s", sig)
	case err := <-errCh:
		if err != nil {
			log.Printf("Bot出错，请检查错误：%v", err)
		}
	}

	log.Printf("Bot已退出，时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
