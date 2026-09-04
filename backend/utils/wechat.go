package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// wechatJsCode2SessionURL 微信 jscode2session 端点。
//
// 暴露为 var（而非 const）便于单元测试用 httptest server 替换：
//   func TestX(t *testing.T) {
//       orig := wechatJsCode2SessionURL
//       wechatJsCode2SessionURL = srv.URL
//       defer func() { wechatJsCode2SessionURL = orig }()
//       ...
//   }
var wechatJsCode2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// JsCode2SessionResponse 微信 jscode2session 接口响应结构。
//
// 官方文档：
// https://developers.weixin.qq.com/miniprogram/dev/api-backend/open-api/login/auth.code2Session.html
type JsCode2SessionResponse struct {
	OpenID      string `json:"openid"`       // 用户唯一标识
	SessionKey  string `json:"session_key"` // 会话密钥
	UnionID     string `json:"unionid"`      // 仅当小程序绑定到微信开放平台时返回
	ErrCode     int    `json:"errcode"`     // 0 表示成功，非 0 见 errmsg
	ErrMsg      string `json:"errmsg"`
}

// JsCode2Session 调用微信小程序 jscode2session，用 code 换 openid 与 session_key。
//
// 刻意用标准库 net/http 而非第三方 SDK，遵循 AGENTS.md「优先选择可跨项目复用的通用
// 依赖」判据——jscode2session 是单次 GET + JSON 解析，引入 SDK 反而把基座业务依赖变重。
// 详见 wechat-mp-login spec 的「jscode2session 调用零第三方依赖」要求。
func JsCode2Session(appid, secret, code string) (string, string, error) {
	return jsCode2SessionDo(wechatJsCode2SessionURL, appid, secret, code, http.DefaultClient)
}

// jsCode2SessionDo 与 JsCode2Session 同语义，但允许注入 baseURL 与 http.Client，便于测试。
//
// baseURL / client 注入是为了让 httptest server 能劫持请求——生产入口 JsCode2Session
// 用全局 wechatJsCode2SessionURL 与 http.DefaultClient，无需关心这两个参数。
func jsCode2SessionDo(baseURL, appid, secret, code string, client *http.Client) (string, string, error) {
	if appid == "" || secret == "" || code == "" {
		return "", "", errors.New("appid/secret/code 不能为空")
	}

	u := baseURL + "?" + url.Values{
		"appid":      {appid},
		"secret":     {secret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}.Encode()

	resp, err := client.Get(u)
	if err != nil {
		return "", "", fmt.Errorf("调用微信 jscode2session 网络错误: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取微信响应失败: %w", err)
	}

	var r JsCode2SessionResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", "", fmt.Errorf("解析微信响应失败: %w（原始响应: %s）", err, string(body))
	}

	// 微信侧错误：errcode 非 0 时（如 code 失效返回 40029），调用方据此响应 400
	if r.ErrCode != 0 {
		return "", "", fmt.Errorf("微信登录失败：errcode=%d errmsg=%s", r.ErrCode, r.ErrMsg)
	}

	// 极端兜底：errcode=0 但 openid 缺失（响应体被截断或非预期格式）
	if r.OpenID == "" {
		return "", "", errors.New("微信响应缺少 openid（原始响应: " + string(body) + ")")
	}

	return r.OpenID, r.SessionKey, nil
}
