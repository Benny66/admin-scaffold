package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestJsCode2Session_Success 微信侧返回 errcode=0 且含 openid：函数返回 openid 与 session_key。
func TestJsCode2Session_Success(t *testing.T) {
	var gotAppID, gotSecret, gotCode, gotGrantType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		gotAppID = q.Get("appid")
		gotSecret = q.Get("secret")
		gotCode = q.Get("js_code")
		gotGrantType = q.Get("grant_type")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"openid":"oABC123","session_key":"sess-key","errcode":0}`))
	}))
	defer srv.Close()

	openid, sessionKey, err := jsCode2SessionDo(srv.URL, "wxtest", "secret", "valid-code", srv.Client())
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if openid != "oABC123" {
		t.Errorf("openid = %q, want oABC123", openid)
	}
	if sessionKey != "sess-key" {
		t.Errorf("session_key = %q, want sess-key", sessionKey)
	}
	// 校验 query 参数被正确拼到 URL 上
	if gotAppID != "wxtest" {
		t.Errorf("appid query = %q, want wxtest", gotAppID)
	}
	if gotSecret != "secret" {
		t.Errorf("secret query = %q, want secret", gotSecret)
	}
	if gotCode != "valid-code" {
		t.Errorf("js_code query = %q, want valid-code", gotCode)
	}
	if gotGrantType != "authorization_code" {
		t.Errorf("grant_type query = %q, want authorization_code", gotGrantType)
	}
}

// TestJsCode2Session_ErrCode 微信侧 errcode != 0：函数返回包含 errcode 的错误。
func TestJsCode2Session_ErrCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()

	_, _, err := jsCode2SessionDo(srv.URL, "wxtest", "secret", "expired-code", srv.Client())
	if err == nil {
		t.Fatal("expected error for errcode != 0, got nil")
	}
	if !strings.Contains(err.Error(), "errcode=40029") {
		t.Errorf("error %q should contain %q", err.Error(), "errcode=40029")
	}
	if !strings.Contains(err.Error(), "invalid code") {
		t.Errorf("error %q should contain errmsg", err.Error())
	}
}

// TestJsCode2Session_EmptyParams 任一参数为空：函数直接返回错误，不发起网络请求。
func TestJsCode2Session_EmptyParams(t *testing.T) {
	cases := []struct{ name, appid, secret, code string }{
		{"empty appid", "", "secret", "code"},
		{"empty secret", "wxtest", "", "code"},
		{"empty code", "wxtest", "secret", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := jsCode2SessionDo("http://x", tc.appid, tc.secret, tc.code, http.DefaultClient)
			if err == nil {
				t.Fatal("expected error for empty param, got nil")
			}
			if !strings.Contains(err.Error(), "appid/secret/code") {
				t.Errorf("error %q should mention empty param", err.Error())
			}
		})
	}
}

// TestJsCode2Session_NetworkError 连接被拒绝：函数返回包含"网络错误"的错误。
func TestJsCode2Session_NetworkError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close() // 关闭 server 使后续请求连接被拒绝

	_, _, err := jsCode2SessionDo(srv.URL, "wxtest", "secret", "code", srv.Client())
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "网络错误") {
		t.Errorf("error %q should mention network error", err.Error())
	}
}

// TestJsCode2Session_MissingOpenID errcode=0 但响应缺 openid：函数返回包含"openid"的错误。
func TestJsCode2Session_MissingOpenID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()

	_, _, err := jsCode2SessionDo(srv.URL, "wxtest", "secret", "code", srv.Client())
	if err == nil {
		t.Fatal("expected error for missing openid, got nil")
	}
	if !strings.Contains(err.Error(), "openid") {
		t.Errorf("error %q should mention missing openid", err.Error())
	}
}

// TestJsCode2Session_MalformedJSON 响应非合法 JSON：函数返回解析错误。
func TestJsCode2Session_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	_, _, err := jsCode2SessionDo(srv.URL, "wxtest", "secret", "code", srv.Client())
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "解析微信响应失败") {
		t.Errorf("error %q should mention parse failure", err.Error())
	}
}
