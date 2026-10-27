package service

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type IdentityService struct{ Client *http.Client }

func NewIdentityService() *IdentityService {
	return &IdentityService{Client: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}}
}
func (s *IdentityService) Verify(fields map[string]string, studentNo, name string) (string, error) {
	if strings.TrimSpace(fields["realname_api_url"]) == "" {
		return "", Reject(503, "站点暂未启用实名认证，请联系管理员")
	}
	resp, e := s.Client.Get(realNameURL(fields["realname_api_url"], studentNo, name))
	if e != nil {
		return "", Reject(502, "实名服务暂不可用，请稍后重试")
	}
	defer resp.Body.Close()
	var result any
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 65536))
	decoder.UseNumber()
	if resp.StatusCode != 200 || decoder.Decode(&result) != nil {
		return "", Reject(502, "实名服务响应异常")
	}
	success, found := jsonField(result, fields["realname_success_field"])
	if !found {
		return "", Reject(502, "实名服务响应缺少配置的成功字段")
	}
	if !jsonMatches(success, fields["realname_success_value"]) {
		return "", Reject(400, "学号与姓名验证未通过")
	}
	idValue, hasID := jsonField(result, fields["realname_id_field"])
	nameValue, hasName := jsonField(result, fields["realname_name_field"])
	returnedID, validID := identityText(idValue)
	realName, validName := identityText(nameValue)
	if !hasID || !hasName || !validID || !validName {
		return "", Reject(502, "实名服务响应缺少配置的学号或姓名字段")
	}
	if returnedID != studentNo || realName != name {
		return "", Reject(400, "学号与姓名不匹配")
	}

	return realName, nil
}
