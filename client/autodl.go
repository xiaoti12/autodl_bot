package client

import (
	"autodl_bot/models"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/go-resty/resty/v2"
)

const (
	BaseURL      = "https://www.autodl.com/api/v1"
	LoginPATH    = "/new_login"
	PassportPath = "/passport"
	InstancePath = "/instance"
)

type AutoDLClient struct {
	client     *resty.Client
	token      string
	tokenMutex sync.RWMutex
	username   string
	password   string
}

func NewAutoDLClient(username, password string) *AutoDLClient {
	client := resty.New()
	client.SetBaseURL(BaseURL)
	client.SetHeaders(map[string]string{
		"accept":             "*/*",
		"accept-language":    "zh-CN,zh;q=0.9",
		"appversion":         "v5.56.0",
		"content-type":       "application/json;charset=UTF-8",
		"sec-ch-ua":          "\"Chromium\";v=\"130\", \"Google Chrome\";v=\"130\", \"Not?A_Brand\";v=\"99\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"Windows\"",
	})

	return &AutoDLClient{
		client:   client,
		username: username,
		password: password,
	}
}

func (c *AutoDLClient) getToken() string {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.token
}
func (c *AutoDLClient) setToken(token string) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.token = token
}
func (c *AutoDLClient) hashPassword(password string) string {
	hasher := sha1.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}
func (c *AutoDLClient) Login() error {
	loginReqest := models.LoginRequest{
		Phone:     c.username,
		Password:  c.hashPassword(c.password),
		VCode:     "",
		PhoneArea: "+86",
		PictureID: nil,
	}
	var loginResponse models.LoginResponse
	_, err := c.client.R().
		SetBody(loginReqest).
		SetResult(&loginResponse).
		Post(LoginPATH)
	if err != nil {
		return err
	}
	if loginResponse.Code != "Success" {
		return errors.New(loginResponse.Msg)
	}

	// get token
	passportRequest := models.PassportRequest{
		Ticket: loginResponse.Data.Ticket,
	}
	var passportResponse models.PassportResponse
	_, err = c.client.R().
		SetBody(passportRequest).
		SetResult(&passportResponse).
		Post(PassportPath)
	if err != nil {
		return err
	}
	if passportResponse.Code != "Success" {
		return errors.New(passportResponse.Msg)
	}
	c.setToken(passportResponse.Data.Token)
	return nil
}
