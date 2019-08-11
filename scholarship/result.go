package scholarship

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetScholarshipResults(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/GATE/SAM/SCHOLARSHIP/S/SJHS06S.ASPX?&maincd=O&systemcd=S&seq=1", consts.ForestURL)
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

	notAvailable := doc.Find("#lblTitle").Text()
	if notAvailable != "" {
		c.String(http.StatusNoContent,
			`It's not the period for checking scholarship results yet.
		장학금 신청 결과 조회 기간이 아닙니다.`)
	}

	results := []gin.H{}
	doc.Find("table#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		results = append(results, gin.H{
			"year":     item.Children().Eq(0).Text(),
			"semester": item.Children().Eq(1).Text(),
			"date":     item.Children().Eq(2).Text(),
			"type":     item.Children().Eq(3).Text(),
			"reason":   item.Children().Eq(4).Text(),
			"result":   item.Children().Eq(5).Text(),
		})
	})

	userinfo := gin.H{
		"univtype": doc.Find("#lblDaehagNm").Text(),
		"depart":   doc.Find("#lblHagbuNm").Text(),
		"major":    doc.Find("#lblSosogNm").Text(),
		"course":   doc.Find("#lblGwajeongNm").Text(),
		"id":       doc.Find("#lblHagbeon").Text(),
		"name":     doc.Find("#lblNm").Text(),
		"status":   doc.Find("#lblHagjeogGbNm").Text(),
		"phone":    doc.Find("#lblHdpNo").Text(),
	}

	c.JSON(http.StatusOK, gin.H{
		"userinfo":      userinfo,
		"apply_results": results,
	})

}
