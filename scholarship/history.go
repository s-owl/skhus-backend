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

	history := []gin.H{}
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
	c.JSON(http.StatusOK, gin.H{
		"scholarship_history": history,
	})
}
