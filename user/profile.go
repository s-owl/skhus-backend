package user

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetUserProfile(c *gin.Context) {
	targetURL := consts.ForestURL + "/GATE/SAM/SERVICE/S/SWSS01P.ASPX?&maincd=O&systemcd=S&seq=1"

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
	imgURL := consts.ForestURL + doc.Find("img#imgSajin").AttrOr("src", "")
	imageData := "data:image/jpeg;base64,"
	if imgURL == consts.ForestURL {
		imageData = ""
	} else {

		client := &http.Client{}
		req, _ := http.NewRequest("GET", imgURL, nil)
		req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
		res, err := client.Do(req)
		if err != nil {
			c.String(http.StatusInternalServerError, consts.InternalError)
			return
		}
		defer res.Body.Close()
		bytes, _ := ioutil.ReadAll(res.Body)
		imageData += base64.StdEncoding.EncodeToString(bytes)
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
		"image":      imageData,
	})
}
