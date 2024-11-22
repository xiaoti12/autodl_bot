package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"autodl_bot/models"

	"github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) (*httptest.Server, *AutoDLClient) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/new_login":
			handleLogin(t, w, r)
		case "/passport":
			handlePassport(t, w, r)
		case "/instance":
			handleInstance(t, w, r)
		default:
			http.NotFound(w, r)
		}
	}))

	// 创建测试客户端
	client := NewAutoDLClient("testuser", "testpass")
	client.client.SetBaseURL(server.URL) // 替换为测试服务器URL

	return server, client
}

func handleLogin(t *testing.T, w http.ResponseWriter, r *http.Request) {
	assert.Equal(t, "POST", r.Method)

	var loginReq models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	assert.NoError(t, err)

	// 验证请求数据
	assert.Equal(t, "testuser", loginReq.Phone)
	assert.NotEmpty(t, loginReq.Password)

	// 返回模拟响应
	resp := map[string]interface{}{
		"code": "Success",
		"msg":  "Success",
		"data": map[string]interface{}{
			"ticket": "test-ticket",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	assert.NoError(t, err)
}

func handlePassport(t *testing.T, w http.ResponseWriter, r *http.Request) {
	assert.Equal(t, "POST", r.Method)

	var passportReq models.PassportRequest
	err := json.NewDecoder(r.Body).Decode(&passportReq)
	assert.NoError(t, err)

	// 验证请求数据
	assert.Equal(t, "test-ticket", passportReq.Ticket)

	// 返回模拟响应
	resp := map[string]interface{}{
		"code": "Success",
		"msg":  "Success",
		"data": map[string]interface{}{
			"token": "test-token",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	assert.NoError(t, err)
}

func handleInstance(t *testing.T, w http.ResponseWriter, r *http.Request) {
	assert.Equal(t, "POST", r.Method)
	assert.Equal(t, "test-token", r.Header.Get("authorization"))

	var instanceReq models.InstanceRequest
	err := json.NewDecoder(r.Body).Decode(&instanceReq)
	assert.NoError(t, err)

	// 返回模拟响应
	resp := map[string]interface{}{
		"code": "Success",
		"msg":  "Success",
		"data": map[string]interface{}{
			"list": []map[string]interface{}{
				{
					"machine_alias": "test-machine",
					"region_name":   "test-region",
					"gpu_all_num":   4,
					"gpu_idle_num":  2,
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	assert.NoError(t, err)
}

func TestLogin(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	err := client.Login()
	assert.NoError(t, err)

	token := client.getToken()
	assert.Equal(t, "test-token", token)

	t.Logf("Login completed, token: %s", token)
}

func TestGetInstances(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	// 测试未登录状态获取实例
	instances, err := client.GetInstances()
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "test-machine", instances[0].MachineAlias)
	assert.Equal(t, "test-region", instances[0].RegionName)
	assert.Equal(t, 4, instances[0].GpuAllNum)
	assert.Equal(t, 2, instances[0].GpuIdleNum)

	// 测试已登录状态获取实例
	client.setToken("test-token")
	instances, err = client.GetInstances()
	assert.NoError(t, err)
	assert.Len(t, instances, 1)

	t.Logf("Get instances completed, instances: %v", instances)
}

func TestGetGPUStatus(t *testing.T) {
	server, client := setupTestServer(t)
	defer server.Close()

	status, err := client.GetGPUStatus()
	assert.NoError(t, err)
	assert.Contains(t, status, "test-machine")
	assert.Contains(t, status, "test-region")
}

func TestHashPassword(t *testing.T) {
	client := NewAutoDLClient("testuser", "testpass")
	hash := client.hashPassword("testpass")
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 40) // SHA1 hash length
	assert.Equal(t, hash, "206c80413b9a96c1312cc346b7d2517b84463edd")
}

func TestAuthorizeFailedRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/instance":
			if r.Header.Get("authorization") == "invalid-token" {
				json.NewEncoder(w).Encode(models.InstanceResponse{
					Code: "AuthorizeFailed",
					Msg:  "Authorization Failed",
				})
			} else {
				handleInstance(t, w, r)
			}
		default:
			// 处理登录和passport请求
			if r.URL.Path == "/new_login" {
				handleLogin(t, w, r)
			} else if r.URL.Path == "/passport" {
				handlePassport(t, w, r)
			}
		}
	}))
	defer server.Close()

	client := NewAutoDLClient("testuser", "testpass")
	client.client.SetBaseURL(server.URL)
	client.setToken("invalid-token")

	// 测试token失效后的自动重试
	instances, err := client.GetInstances()
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "test-token", client.getToken())
}
