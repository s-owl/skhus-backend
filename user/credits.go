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

func GetMyCredits(c *gin.Context) {
	credential := c.GetHeader("CredentialOld")
	_, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}
	targetURL := fmt.Sprintf("%s/Gate/UniMainStudent.aspx", consts.ForestURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", credential)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}
	creditsData := []gin.H{}
	items := doc.Find("#divContainer > div:nth-child(4) > table > tbody > tr")
	for i := 0; i < 14; i += 2 {
		for j := 0; j < 3; j++ {
			if i == 10 && j != 0 {
				break
			}
			creditsData = append(creditsData, gin.H{
				"type":   strings.TrimSpace(items.Eq(i).Children().Eq(j).Text()),
				"earned": strings.TrimSpace(items.Eq(i + 1).Children().Eq(j).Text()),
			})
		}
	}
	summary := doc.Find("span#CORX03C1_lblTot").Text()
	c.JSON(http.StatusOK, gin.H{
		"credits": creditsData,
		"summary": summary,
	})
}
