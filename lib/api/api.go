package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"modstart-cli/lib/config"
	"modstart-cli/lib/version"
)

// ErrNeedLogin 凭证失效
var ErrNeedLogin = errors.New("凭证已失效，请重新执行 modstart-cli login")

// Response 后台统一响应
type Response struct {
	Code     int             `json:"code"`
	Msg      string          `json:"msg"`
	Data     json.RawMessage `json:"data"`
	Redirect string          `json:"redirect"`
}

// View 转换为便于 JSON 输出的结构（data 解析为普通对象，避免中文被转义）
func (r *Response) View() map[string]interface{} {
	var data interface{}
	if len(r.Data) > 0 {
		_ = json.Unmarshal(r.Data, &data)
	}
	view := map[string]interface{}{
		"code": r.Code,
		"msg":  r.Msg,
		"data": data,
	}
	if r.Redirect != "" {
		view["redirect"] = r.Redirect
	}
	return view
}

// ParamDef 接口参数描述
type ParamDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

// ApiDef 接口描述
type ApiDef struct {
	Key    string     `json:"key"`
	Title  string     `json:"title"`
	Method string     `json:"method"`
	URL    string     `json:"url"`
	Scope  string     `json:"scope"`
	Params []ParamDef `json:"params"`
	Module string     `json:"module"`
	App    string     `json:"app"`
}

// ApiListData api_list 返回数据
type ApiListData struct {
	AdminPrefix string   `json:"adminPrefix"`
	Apis        []ApiDef `json:"apis"`
}

// Client 站点客户端
type Client struct {
	Site *config.Site
	HTTP *http.Client
}

// New 创建客户端
func New(site *config.Site) *Client {
	return &Client{
		Site: site,
		HTTP: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) app() string {
	if c.Site.App == "" {
		return "admin"
	}
	return c.Site.App
}

func (c *Client) adminPath() string {
	adminPath := strings.Trim(c.Site.AdminPath, "/")
	if adminPath == "" {
		return "admin"
	}
	return adminPath
}

// cliBase CLI 路由前缀：后台固定为 cli，其他身份为 <app>_cli
func (c *Client) cliBase() string {
	app := c.app()
	if app == "" || app == "admin" {
		return "cli"
	}
	return app + "_cli"
}

// cliURL CLI 接口地址
func (c *Client) cliURL(path string) string {
	return c.Site.URL + "/" + c.adminPath() + "/" + c.cliBase() + "/" + strings.TrimLeft(path, "/")
}

func (c *Client) do(method, rawURL string, params map[string]string, extraHeaders map[string]string) (*http.Response, []byte, error) {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	var body io.Reader
	if method == http.MethodPost {
		body = bytes.NewBufferString(values.Encode())
	} else if encoded := values.Encode(); encoded != "" {
		rawURL = rawURL + "?" + encoded
	}

	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, nil, err
	}
	// CLI 凭证每次请求都通过请求头传递，服务端不接受请求体/URL 凭证
	if c.Site.Token != "" {
		req.Header.Set("admin-cli-key", c.Site.Token)
	}
	req.Header.Set("User-Agent", "modstart-cli/"+version.VERSION)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, data, nil
}

func parse(data []byte) (*Response, error) {
	var r Response
	if err := json.Unmarshal(data, &r); err != nil {
		preview := string(data)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		return nil, fmt.Errorf("返回内容不是 JSON: %s", preview)
	}
	return &r, nil
}

// Login 校验凭证并保存（无状态，不使用 Cookie / Session）
func (c *Client) Login(token, name string) (*Response, error) {
	if token != "" {
		c.Site.Token = token
	}
	resp, data, err := c.do(http.MethodPost, c.cliURL("me"), nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf(
			"接口不存在（HTTP 404）：请确认 %s 是后台管理地址（如 https://demo.com/admin）、站点已安装 AdminCli 模块",
			c.Site.AdminURL,
		)
	}

	r, err := parse(data)
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		if r.Msg == "" {
			r.Msg = "登录失败"
		}
		return r, errors.New(r.Msg)
	}
	return r, nil
}

func (c *Client) relogin() error {
	if c.Site.Token == "" {
		return ErrNeedLogin
	}
	_, err := c.Login(c.Site.Token, c.Site.Name)
	return err
}

// Call 调用后台原接口
func (c *Client) Call(path string, method string, params map[string]string) (*Response, error) {
	r, err := c.callOnce(path, method, params)
	if err == ErrNeedLogin {
		if reloginErr := c.relogin(); reloginErr != nil {
			return r, ErrNeedLogin
		}
		return c.callOnce(path, method, params)
	}
	return r, err
}

func (c *Client) callOnce(path string, method string, params map[string]string) (*Response, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodGet && method != http.MethodPost {
		method = http.MethodPost
	}
	resp, data, err := c.do(method, c.cliURL(path), params, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := strings.ToLower(resp.Header.Get("Location"))
		if strings.Contains(location, "login") {
			return nil, ErrNeedLogin
		}
	}
	r, err := parse(data)
	if err != nil {
		return nil, err
	}
	if r.Redirect != "" && strings.Contains(strings.ToLower(r.Redirect), "login") {
		return r, ErrNeedLogin
	}
	if r.Code != 0 {
		if r.Msg == "" {
			r.Msg = "调用失败"
		}
		return r, errors.New(r.Msg)
	}
	return r, nil
}

func (c *Client) cliPost(path string, params map[string]string) (*Response, error) {
	_, data, err := c.do(http.MethodPost, c.cliURL(path), params, nil)
	if err != nil {
		return nil, err
	}
	r, err := parse(data)
	if err != nil {
		return nil, err
	}
	if r.Redirect != "" && strings.Contains(strings.ToLower(r.Redirect), "login") {
		return r, ErrNeedLogin
	}
	return r, nil
}

// ApiList 获取当前站点可用的 CLI 接口描述
func (c *Client) ApiList() (*ApiListData, error) {
	r, err := c.cliPost("api_list", nil)
	if err == ErrNeedLogin {
		if reloginErr := c.relogin(); reloginErr != nil {
			return nil, ErrNeedLogin
		}
		r, err = c.cliPost("api_list", nil)
	}
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		if r.Msg == "" {
			r.Msg = "获取接口列表失败"
		}
		return nil, errors.New(r.Msg)
	}
	var out ApiListData
	if err := json.Unmarshal(r.Data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Me 当前登录信息
func (c *Client) Me() (map[string]interface{}, error) {
	r, err := c.cliPost("me", nil)
	if err == ErrNeedLogin {
		if reloginErr := c.relogin(); reloginErr != nil {
			return nil, ErrNeedLogin
		}
		r, err = c.cliPost("me", nil)
	}
	if err != nil {
		return nil, err
	}
	if r.Code != 0 {
		return nil, errors.New(r.Msg)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(r.Data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Logout 退出登录（无状态，仅清除本地凭证）
func (c *Client) Logout() error {
	c.Site.Token = ""
	return nil
}
