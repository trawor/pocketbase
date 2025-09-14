package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	NameWecom = "wecom"
)

func init() {
	Providers[NameWecom] = wrapFactory(NewWecomProvider)
}

type WecomProvider struct {
	BaseProvider
	AgentID     string
	appToken    string
	appTokenExp time.Time
}

func NewWecomProvider() *WecomProvider {
	p := &WecomProvider{}
	p.SetPKCE(false)
	p.SetDisplayName("企业微信")
	// p.SetScopes([]string{"snsapi_privateinfo"})
	p.SetAuthURL("https://login.work.weixin.qq.com/wwlogin/sso/login") //"https://open.weixin.qq.com/connect/oauth2/authorize"
	p.SetTokenURL("https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo")
	p.SetUserInfoURL("https://qyapi.weixin.qq.com/cgi-bin/auth/getuserdetail")
	return p
}

func (p *WecomProvider) FetchToken(code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	fmt.Println("FetchToken, code:", code, opts)
	return &oauth2.Token{
		AccessToken: code,
	}, nil
}

func (p *WecomProvider) FetchRawUserInfo(token *oauth2.Token) ([]byte, error) {
	appToken, err := p.getAppToken()
	if err != nil {
		return nil, err
	}
	rurl := p.tokenURL + "?access_token=" + appToken + "&code=" + token.AccessToken

	req, _ := http.NewRequest("GET", rurl, nil)
	return httpRequest(req)
}

func (p *WecomProvider) FetchAuthUser(token *oauth2.Token) (user *AuthUser, err error) {
	body, err := p.FetchRawUserInfo(token)
	// fmt.Println("FetchAuthUser", "body", string(body), "error", err)
	if err != nil {
		return nil, err
	}
	info := struct {
		Code        int64  `json:"errcode"`
		Error       string `json:"errmsg"`
		Userid      string `json:"userid"`
		User_ticket string `json:"user_ticket"`

		Openid          string `json:"openid"`
		External_userid string `json:"external_userid"`
	}{}
	json.Unmarshal(body, &info)

	if info.Code != 0 || info.Error != "ok" {
		return nil, errors.New(info.Error)
	}

	if info.User_ticket != "" {
		// 可以获取详细资料
		user, err = p.getUserInfo(info.User_ticket)
		if err != nil {
			return nil, err
		}
	} else {
		user = &AuthUser{
			Username: info.Userid,
		}
	}

	durl := "https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=" + p.appToken + "&userid=" + info.Userid
	req, _ := http.NewRequest("GET", durl, nil)

	body, err = httpRequest(req)
	if err != nil {
		return nil, err
	}
	var ret map[string]any
	if err := json.Unmarshal(body, &ret); err != nil {
		return nil, err
	}
	if ret["errmsg"] != "ok" {
		return nil, errors.New(ret["errmsg"].(string))
	}
	if ret["status"].(float64) != 1 {
		return nil, fmt.Errorf("user status=%v, no access", ret["status"])
	}

	if ret["name"] != "" {
		user.Name = ret["name"].(string)
	}

	if info.Openid != "" {
		user.Id = info.Openid
	} else {
		// TODO 换取 openid
	}
	if user.Id == "" {
		user.Id = user.Username
	}
	return user, err
}

func httpRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (p *WecomProvider) BuildAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	url := p.authURL + "?response_type=code" +
		"&appid=" + p.ClientId() +
		"&state=" + state +
		"&login_type=CorpApp"
	// "#wechat_redirect"
	if agentId, ok := p.Extra()["AgentID"]; ok {
		url += "&scope=snsapi_privateinfo&agentid=" + agentId.(string)
	} else {
		url += "&scope=snsapi_base"
	}
	return url
}

func (p *WecomProvider) getAppToken() (string, error) {
	// Share token in app store
	if p.appToken != "" && time.Since(p.appTokenExp) < -10*time.Second {
		return p.appToken, nil
	}

	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" +
		p.ClientId() +
		"&corpsecret=" +
		p.ClientSecret()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	body, err := httpRequest(req)
	if err != nil {
		return "", err
	}

	tokenInfo := struct {
		Code             int64  `json:"errcode"`
		Expire           int64  `json:"expires_in"`
		App_access_token string `json:"access_token"`
	}{}
	json.Unmarshal(body, &tokenInfo)

	if tokenInfo.Code != 0 {
		return "", errors.New(string(body))
	}
	slog.Debug("getAppToken", "token", tokenInfo.App_access_token)
	p.appToken = tokenInfo.App_access_token
	p.appTokenExp = time.Now().Add(time.Duration(tokenInfo.Expire) * time.Second)
	return p.appToken, nil
}

func (p *WecomProvider) getUserInfo(ticket string) (*AuthUser, error) {
	appToken, err := p.getAppToken()
	if err != nil {
		return nil, err
	}

	rurl := p.UserInfoURL() + "?access_token=" + appToken
	info := `{"user_ticket":"` + ticket + `"}`
	req, _ := http.NewRequest("POST", rurl, bytes.NewBuffer([]byte(info)))

	body, err := httpRequest(req)
	if err != nil {
		return nil, err
	}
	slog.Debug("getUserDetail", "url", rurl, "ticket", info)

	/*
			{
		   "errcode": 0,
		   "errmsg": "ok",
		   "userid":"lisi",
		   "gender":"1",
		   "avatar":"http://shp.qpic.cn/bizmp/xxxxxxxxxxx/0",
		   "qr_code":"https://open.work.weixin.qq.com/wwopen/userQRCode?vcode=vcfc13b01dfs78e981c",
		   "mobile": "13800000000",
		   "email": "zhangsan@gzdev.com",
		   "biz_mail":"zhangsan@qyycs2.wecom.work",
		   "address": "广州市海珠区新港中路"
		}
	*/
	var ret map[string]any
	if err := json.Unmarshal(body, &ret); err != nil {
		return nil, err
	}
	if ret["errmsg"] != "ok" {
		return nil, errors.New(ret["errmsg"].(string))
	}
	delete(ret, "errcode")
	delete(ret, "errmsg")
	user := &AuthUser{
		Username:  ret["userid"].(string),
		AvatarURL: ret["avatar"].(string),
	}
	if email := ret["email"].(string); email != "" {
		user.Email = email
	} else {
		if email := ret["biz_mail"].(string); email != "" {
			user.Email = email
		}
		delete(ret, "email")
	}
	user.RawUser = ret
	slog.Debug("getUserInfo", "info", ret)
	return user, nil
}
