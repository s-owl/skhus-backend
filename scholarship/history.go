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

func GetScholarshipHistory(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/GATE/SAM/SCHOLARSHIP/S/SJHS01S.ASPX?&maincd=O&systemcd=S&seq=1", consts.ForestURL)
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

	history := []gin.H{}
	doc.Find("table#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		history = append(history, gin.H{
			"year":             item.Children().Eq(0).Text(),
			"semester":         item.Children().Eq(1).Text(),
			"scholarship_name": item.Children().Eq(2).Text(),
			"order":            item.Children().Eq(3).Text(),
			"grade":            item.Children().Eq(4).Text(),
			"amount_entrance":  item.Children().Eq(5).Text(),
			"amount_class":     item.Children().Eq(6).Text(),
			"benefit_type":     item.Children().Eq(7).Text(),
			"note":             item.Children().Eq(8).Text(),
		})
	})
	c.JSON(http.StatusOK, gin.H{
		"scholarship_history": history,
	})
}
