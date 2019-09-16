package user

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetClassInfo(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/Gate/UniMainStudent.aspx", consts.ForestURL)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
	res, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, consts.InternalError)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"counselor": doc.Find("span#lblCounselor").Text(),
		"class":     doc.Find("span#lblBanBunBan").Text(),
		"google":    doc.Find("span#lblGoogleClassAccount").Text(),
	})
}
