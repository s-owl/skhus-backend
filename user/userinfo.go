package user

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetUserinfo(c *gin.Context) {
	credential := c.GetHeader("CredentialOld")
	cookies, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}
	targetURL := fmt.Sprintf("%s/Gate/UniTopMenu.aspx", consts.ForestURL)
	cr := colly.NewCollector()
	cr.OnHTML("span#lblInfo", func(e *colly.HTMLElement) {
		splited := strings.Split(e.Text, ":")
		userinfo := strings.Split(splited[1], "(")
		name := userinfo[0]
		id := strings.Split(userinfo[1], ")")[0]
		c.JSON(http.StatusOK, gin.H{
			"userinfo": gin.H{
				"name": name,
				"id":   id,
			},
		})
	})
	cr.SetCookies(consts.ForestURL, cookies)
	cr.Visit(targetURL)
}
