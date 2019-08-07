package tools

import (
	"errors"
	"net/http"
	"strings"
)

var CookieConvertError error = errors.New(
	"tools.ConvertToCookies : cookies string is empty or not in correct form")

func ConvertToCookies(cookies string) ([]*http.Cookie, error) {
	if !strings.Contains(cookies, ";") || !strings.Contains(cookies, "=") || cookies == "" {
		return nil, CookieConvertError
	}

	header := http.Header{}
	header.Add("Cookie", cookies)
	request := http.Request{Header: header}
	return request.Cookies(), nil
}
