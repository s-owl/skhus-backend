package tools

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
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

func EucKrReaderToUtf8Reader(body io.Reader) io.Reader {
	rInUTF8 := transform.NewReader(body, korean.EUCKR.NewDecoder())
	decBytes, _ := ioutil.ReadAll(rInUTF8)
	decrypted := string(decBytes)
	return strings.NewReader(decrypted)
}
