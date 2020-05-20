package grade

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

// Forest에서 학점을 가져온다.
/*
	{
	}
*/
func GetGradeCertificate(c *gin.Context) {
	// forest에 접근해서 html을 받아온다.
	targetURL := consts.ForestURL + "/GATE/SAM/SCORE/S/SSJS06S.ASPX?&maincd=O&systemcd=S&seq=1"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
	res, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, consts.InternalError)
		log.Println(err)
		return
	}
	defer res.Body.Close()

	// 받아온 html을 goquery로 분석한다.
	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Println(err)
		return
	}

	// 사용자의 정보를 담을 변수를 초기화
	userinfo := []gin.H{}
	// 사용자의 정보가 담긴 테이블을 찾는다.
	userdata := doc.Find("#Table3 > tbody > tr > td")
	// 테이블을 순회해서 userinfo에 담는다.
	userdata.Each(func(i int, item *goquery.Selection) {
		if i%2 != 0 {
			userinfo = append(userinfo, gin.H{
				"name":  userdata.Eq(i - 1).Text(),
				"value": strings.TrimSpace(item.Text()),
			})
		}
	})

	// 사용자의 학점을 담을 변수를 초기화
	details := []gin.H{}
	// 학점이 존재하는 테이블을 찾아 순회한다.
	doc.Find("table#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		// 넌도, 학기, 교유번호, 과목, 종류, 학점, 등급 순으로 들어온다.
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

	// 요약 부분을 저장할 변수를 초기화한다.
	summary := []gin.H{}
	// 요약 정보를 찾고 순회해서 저장한다.
	summaryData := doc.Find("table#Table2 > tbody > tr")
	for i := 0; i < 17; i++ {
		summary = append(summary, gin.H{
			"type":   summaryData.Eq(0).Children().Eq(i).Text(),
			"credit": summaryData.Eq(1).Children().Eq(i).Text(),
		})
	}

	// 인증 날짜를 찾아 저장한다.
	certDate := doc.Find("#lblDt").Text()

	// 결과값을 전송한다.
	c.JSON(http.StatusOK, gin.H{
		"userinfo": userinfo,
		"details":  details,
		"summary":  summary,
		"date":     certDate,
	})
}
