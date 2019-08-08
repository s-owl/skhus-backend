package user

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetUserinfo(c *gin.Context) {
	credential := c.GetHeader("CredentialOld")
	_, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}
	targetURL := fmt.Sprintf("%s/Gate/UniTopMenu.aspx", consts.ForestURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", credential)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}
	e := doc.Find("span#lblInfo")
	splited := strings.Split(e.Text(), ":")
	userinfo := strings.Split(splited[1], "(")
	name := userinfo[0]
	id := strings.Split(userinfo[1], ")")[0]
	c.JSON(http.StatusOK, gin.H{
		"userinfo": gin.H{
			"name": name,
			"id":   id,
		},
	})
}
