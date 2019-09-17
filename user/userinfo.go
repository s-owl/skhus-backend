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
	targetURL := fmt.Sprintf("%s/Gate/UniTopMenu.aspx", consts.ForestURL)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
	res, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, consts.InternalError)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Println(err)
		return
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
