package enroll

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

func GetSavedCredits(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/Gate/SAM/Lecture/H/SSGH03S.aspx?&maincd=O&systemcd=S&seq=100", consts.ForestURL)
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
