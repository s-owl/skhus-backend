package scholarship

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

// 학생의 장학내역을 받아오는 기능을 구현
/*
	{
		"scholarship_history": [
			{
				"year":             "...",
				"semester":         "...",
				"scholarship_name": "...",
				"amount_entrance":  "...",
				"amount_class":     "...",
				"benefit_type":     "...",
				"note":             "...",
			}, ...
		]
	}
*/
func GetScholarshipHistory(c *gin.Context) {
	// Forest에서 장학 기록을 받아온다.
	targetURL := consts.ForestURL + "/GATE/SAM/SCHOLARSHIP/S/SJHS01S.ASPX?&maincd=O&systemcd=S&seq=1"
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

	// 결과값을 담기 위한 변수를 초기화한다.
	history := []gin.H{}
	// 테이블을 찾아 장학 기록을 history 변수에 담는다.
	doc.Find("table#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		history = append(history, gin.H{
			"year":             item.Children().Eq(0).Text(),
			"semester":         item.Children().Eq(1).Text(),
			"scholarship_name": item.Children().Eq(2).Text(),
			"amount_entrance":  item.Children().Eq(3).Text(),
			"amount_class":     item.Children().Eq(4).Text(),
			"benefit_type":     item.Children().Eq(5).Text(),
			"note":             item.Children().Eq(6).Text(),
		})
	})
	// 결과값을 전송한다.
	c.JSON(http.StatusOK, gin.H{
		"scholarship_history": history,
	})
}
