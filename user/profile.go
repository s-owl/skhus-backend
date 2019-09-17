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

func GetUserProfile(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/GATE/SAM/SERVICE/S/SWSS01P.ASPX?&maincd=O&systemcd=S&seq=1", consts.ForestURL)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
	res, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, consts.InternalError)
		log.Fatal(err)
		return
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(tools.EucKrReaderToUtf8Reader(res.Body))
	if err != nil {
		log.Fatal(err)
	}
	imgURL := consts.ForestURL + doc.Find("img#imgSajin").AttrOr("src", "")
	if imgURL == consts.ForestURL {
		imgURL = ""
	}
	c.JSON(http.StatusOK, gin.H{
		"name":       doc.Find("span#lblNm").Text(),
		"id":         doc.Find("span#lblHagbeon").Text(),
		"college":    doc.Find("span#lblDaehagCdNm").Text(),
		"department": doc.Find("span#lblHagbuCdNm").Text(),
		"major":      doc.Find("span#lblSosogCdNm").Text(),
		"grade":      doc.Find("span#lblHagnyeon").Text(),
		"classtype":  doc.Find("span#lblJuyaGbNm").Text(),
		"coursetype": doc.Find("span#lblGwajeongGbNm").Text(),
		"state":      doc.Find("span#lblHagjeogStGbNm").Text(),
		"image":      imgURL,
	})
}
