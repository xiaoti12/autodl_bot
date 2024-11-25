package client

import (
	"autodl_bot/models"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	BaseURL      = "https://www.autodl.com/api/v1"
	LoginPATH    = "/new_login"
	PassportPath = "/passport"
	InstancePath = "/instance"
	PowerOnPath  = "/instance/power_on"
	PowerOffPath = "/instance/power_off"
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

func (c *AutoDLClient) Login() error {
	loginReqest := models.LoginRequest{
		Phone:     c.username,
		Password:  c.password,
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
		log.Printf("[ERROR] 登录请求失败: %v", err)
		return err
	}
	if loginResponse.Code != "Success" {
		log.Printf("[ERROR] 登录失败: %v", err)
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
		log.Printf("[ERROR] 获取 token 请求失败: %v", err)
		return err
	}
	if passportResponse.Code != "Success" {
		log.Printf("[ERROR] 获取 token 失败: %v", err)
		return errors.New(passportResponse.Msg)
	}
	c.setToken(passportResponse.Data.Token)
	log.Printf("[INFO] 用户%s登录成功，获取到token", c.username)
	return nil
}

func (c *AutoDLClient) GetInstances() ([]models.Instance, error) {
	token := c.getToken()
	if token == "" {
		log.Printf("[INFO] 用户%stoken不存在，重新登录", c.username)
		if err := c.Login(); err != nil {
			return nil, err
		}
		token = c.getToken()
	}

	instanceRequest := models.InstanceRequest{
		DateFrom:   "",
		DateTo:     "",
		PageIndex:  1,
		PageSize:   10,
		Status:     []string{},
		ChargeType: []string{},
	}

	var instanceResponse models.InstanceResponse
	_, err := c.client.R().
		SetHeader("authorization", token).
		SetBody(instanceRequest).
		SetResult(&instanceResponse).
		Post(InstancePath)

	if err != nil {
		return nil, err
	}
	// check if token valid
	if instanceResponse.Code == "AuthorizeFailed" {
		// re-login
		log.Printf("[INFO] 用户%s登录过期，重新登录", c.username)
		if err := c.Login(); err != nil {
			return nil, err
		}
		_, err := c.client.R().
			SetHeader("authorization", c.getToken()).
			SetBody(instanceRequest).
			SetResult(&instanceResponse).
			Post(InstancePath)

		if err != nil {
			log.Printf("[ERROR] 查询实例请求失败: %v", err)
			return nil, err
		}
	}

	if instanceResponse.Code != "Success" {
		log.Printf("[ERROR] 查询实例失败: %v", err)
		return nil, errors.New(instanceResponse.Msg)
	}

	log.Printf("[INFO] 用户%s查询实例成功", c.username)
	return instanceResponse.Data.List, nil
}

func (c *AutoDLClient) GetGPUStatus() (string, error) {
	instances, err := c.GetInstances()
	if err != nil {
		return "", err
	}

	var result string
	for i, instance := range instances {
		result += fmt.Sprintf("机器: %s-%s\n", instance.RegionName, instance.MachineAlias)
		result += "UUID: " + instance.UUID + "\n"
		result += fmt.Sprintf("GPU: %d/%d\n", instance.GpuIdleNum, instance.GpuAllNum)
		result += getReleaseTime(instance.StoppedAt.Time)
		if i < len(instances)-1 {
			result += "----------------\n"
		}
	}
	return result, nil
}

func (c *AutoDLClient) PowerOn(uuid string) error {
	if uuid == "" {
		return errors.New("实例UUID不能为空")
	}

	body := map[string]string{
		"instance_uuid": uuid,
	}
	var response models.PowerResponse

	_, err := c.client.R().
		SetHeader("authorization", c.getToken()).
		SetBody(body).
		SetResult(&response).
		Post(PowerOnPath)
	if err != nil {
		return fmt.Errorf("开机请求失败: %v", err)
	}
	if response.Code != "Success" {
		return fmt.Errorf("开机失败: %s", response.Msg)
	}

	log.Printf("[INFO] 用户%s实例 %s 开机成功", c.username, uuid)
	return nil
}

func (c *AutoDLClient) PowerOff(uuid string) error {
	if uuid == "" {
		return errors.New("实例UUID不能为空")
	}

	body := map[string]string{
		"instance_uuid": uuid,
	}
	var response models.PowerResponse
	_, err := c.client.R().
		SetHeader("authorization", c.getToken()).
		SetBody(body).
		SetResult(&response).
		Post(PowerOffPath)
	if err != nil {
		return fmt.Errorf("关机请求失败: %v", err)
	}
	if response.Code != "Success" {
		return fmt.Errorf("关机失败: %s", response.Msg)
	}

	log.Printf("[INFO] 用户%s实例 %s 关机成功", c.username, uuid)
	return nil
}

func HashPassword(rawPassword string) string {
	hasher := sha1.New()
	hasher.Write([]byte(rawPassword))
	return hex.EncodeToString(hasher.Sum(nil))
}

func getReleaseTime(stoppedTime string) string {
	result := "释放时间："
	stoppedAt, err := time.Parse("2006-01-02T15:04:05+08:00", stoppedTime)
	if err != nil {
		return result + "解析失败"
	}
	releaseTime := stoppedAt.Add(15 * 24 * time.Hour).Sub(time.Now())
	if releaseTime > 0 {
		result += fmt.Sprintf("%s后释放\n", formatDuration(releaseTime))
	} else {
		result += "已释放"
	}
	return result
}

func formatDuration(d time.Duration) string {
	days := d / (24 * time.Hour)
	hours := (d % (24 * time.Hour)) / time.Hour
	minutes := (d % time.Hour) / time.Minute

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}
