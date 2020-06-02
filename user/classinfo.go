package user

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetClassInfo(c *gin.Context) {
	targetURL := consts.ForestURL + "/Gate/UniMainStudent.aspx"

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

	c.JSON(http.StatusOK, gin.H{
		"counselor": doc.Find("span#lblCounselor").Text(),
		"class":     doc.Find("span#lblBanBunBan").Text(),
		"google":    doc.Find("span#lblGoogleClassAccount").Text(),
	})
}
