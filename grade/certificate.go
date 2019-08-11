package grade

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

func GetGradeCertificate(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/GATE/SAM/SCORE/S/SSJS06S.ASPX?&maincd=O&systemcd=S&seq=1", consts.ForestURL)
	credential := c.GetHeader("CredentialOld")
	_, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", credential)
	res, err := client.Do(req)
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}

	userinfo := []gin.H{}
	userdata := doc.Find("#Table3 > tbody > tr > td")
	userdata.Each(func(i int, item *goquery.Selection) {
		if i%2 != 0 {
			userinfo = append(userinfo, gin.H{
				"name":  userdata.Eq(i - 1).Text(),
				"value": strings.TrimSpace(item.Text()),
			})
		}
	})

	details := []gin.H{}
	doc.Find("table#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		details = append(details, gin.H{
			"year":     item.Children().Eq(0).Text(),
			"semester": item.Children().Eq(1).Text(),
			"code":     item.Children().Eq(2).Text(),
			"subject":  item.Children().Eq(3).Text(),
			"type":     item.Children().Eq(4).Text(),
			"credit":   item.Children().Eq(5).Text(),
			"grade":    item.Children().Eq(6).Text(),
		})
	})

	summary := []gin.H{}
	summaryData := doc.Find("table#Table2 > tbody > tr")
	for i := 0; i < 17; i++ {
		summary = append(summary, gin.H{
			"type":   summaryData.Eq(0).Children().Eq(i).Text(),
			"credit": summaryData.Eq(1).Children().Eq(i).Text(),
		})
	}

	certDate := doc.Find("#lblDt").Text()

	c.JSON(http.StatusOK, gin.H{
		"userinfo": userinfo,
		"details":  details,
		"summary":  summary,
		"date":     certDate,
	})
}
