package tools

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
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

func CredentialOldCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		credential := c.GetHeader("CredentialOld")
		if credential == "" {
			c.String(http.StatusBadRequest, consts.CredentialMalformedMsg)
		}
		for _, item := range []string{"ASP.NET_SessionId", ".AuthCookie", "UniCookie", "KIS"} {
			if !strings.Contains(credential, item) {
				c.String(http.StatusBadRequest, consts.CredentialMalformedMsg)
			}
		}
		if len(strings.Split(credential, ";")) != 4 {
			c.String(http.StatusBadRequest, consts.CredentialMalformedMsg)
		}
		cookies, err := ConvertToCookies(credential)
		if err != nil {
			c.String(http.StatusBadRequest, consts.CredentialMalformedMsg)
		}
		c.Set("CredentialOldCookies", cookies)
		c.Set("CredentialOld", credential)
		c.Next()
	}
}
