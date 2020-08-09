package enroll

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

/*
학점 세이브 제도로 보관한 학점을 표시한다.
{
	"status": {
		"accrued": "숫자"(발생한 학점),
		"accured_criteria": "년도 학기"(이 학점이 누적된 시점),
		"used": "숫자"(사용한 학점),
		"used_criteria": "년도 학기"(이 학점이 사용된 시점),
		"available": "숫자"(사용할 수 있는 학점)
	},
	"details": [
		{
			"year": "년도",
			"semester": "학기",
			"saved": "숫자"(저장된 학점),
			"used": "숫자"(사용된 학점),
		}, ...
	],
}
*/
func GetSavedCredits(c *gin.Context) {
	targetURL := consts.ForestURL + "/Gate/SAM/Lecture/H/SSGH03S.aspx?&maincd=O&systemcd=S&seq=100"
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

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Println(err)
		return
	}

	details := []gin.H{}
	doc.Find("#gvDetails > tbody > tr").Each(func(i int, item *goquery.Selection) {
		if i > 0 {
			details = append(details, gin.H{
				"year":     item.Children().Eq(0).Text(),
				"semester": item.Children().Eq(1).Text(),
				"saved":    item.Children().Eq(2).Text(),
				"used":     item.Children().Eq(3).Text(),
			})
		}
	})

	status := doc.Find("table#fvList > tbody > tr > td > table.gridForm > tbody > tr > td")

	c.JSON(http.StatusOK, gin.H{
		"status": gin.H{
			"accrued":          strings.TrimSpace(status.Eq(0).Text()),
			"accrued_criteria": strings.TrimSpace(status.Eq(1).Text()),
			"used":             strings.TrimSpace(status.Eq(2).Text()),
			"used_criteria":    strings.TrimSpace(status.Eq(3).Text()),
			"available":        strings.TrimSpace(status.Eq(4).Text()),
		},
		"details": details,
	})

}
