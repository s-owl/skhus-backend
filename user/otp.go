package user

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

func GetOtpCode(c *gin.Context) {
	targetURL := fmt.Sprintf("%s/Gate/Utility/A/UTLA01P.aspx?&maincd=O&systemcd=S&seq=0", consts.ForestURL)

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

	otpCode := doc.Find("span#lblOtpNum").Text()
	timeLeftRaw := doc.Find("span#lblRemainSec").Text()
	timeLeftSec := strings.Count(timeLeftRaw, "■") * 5

	c.JSON(http.StatusOK, gin.H{
		"otpcode":       otpCode,
		"time_left_sec": timeLeftSec,
	})
}
