package user

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

func GetCurrentAttendance(c *gin.Context) {
	credential := c.GetHeader("CredentialOld")
	_, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}
	targetURL := fmt.Sprintf("%s/Gate/UniMainStudent.aspx", consts.ForestURL)

	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", credential)
	res, err := client.Do(req)
	defer res.Body.Close()

	c.JSON(http.StatusOK, extractData(tools.EucKrReaderToUtf8Reader(res.Body)))
}

type AttendanceOption struct {
	Year     string `form:"year" json:"year" xml:"year"  binding:"required"`
	Semester string `form:"semester" json:"semester" xml:"semester"  binding:"required"`
}

func GetAttendanceWithOptions(c *gin.Context) {

	credential := c.GetHeader("CredentialOld")
	cookies, err := tools.ConvertToCookies(credential)
	if err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed credential data.
			비어 있거나 올바르지 않은 인증 데이터 입니다.`)
		return
	}

	var optionData AttendanceOption
	if err := c.ShouldBindJSON(&optionData); err != nil {
		c.String(http.StatusBadRequest,
			`Empty or malformed option data.
			비어 있거나 올바르지 않은 조건 데이터 입니다.`)
		return
	}

	targetURL := fmt.Sprintf("%s/Gate/UniMainStudent.aspx", consts.ForestURL)
	// Options for custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE))

	// Create contexts
	ctx, cancelCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelCtx()

	var content string

	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.ActionFunc(func(context context.Context) error {
			fmt.Println("Setting cookies")
			for _, item := range cookies {
				network.SetCookie(item.Name, item.Value)
			}
			return nil
		}),
		chromedp.Navigate(targetURL),
		chromedp.WaitReady(`txtYy`, chromedp.ByID),
		chromedp.SetValue(`txtYy`, optionData.Year, chromedp.ByID),
		chromedp.SetValue(`ddlHaggi`, optionData.Semester, chromedp.ByID),
		chromedp.Click(`btnList`, chromedp.ByID),
		chromedp.WaitReady(`txtYy`, chromedp.ByID),
		chromedp.InnerHTML(`body`, &content, chromedp.ByQuery),
	})

	c.JSON(http.StatusOK, extractData(strings.NewReader(content)))
}

func extractData(body io.Reader) map[string]interface{} {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}
	attendanceData := []gin.H{}
	doc.Find("#gvList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		subjectStr := strings.TrimSpace(item.Children().Eq(0).Text())
		splitedSubjArr := strings.Split(subjectStr, "(")
		attendanceData = append(attendanceData, gin.H{
			"subject_code": strings.Trim(splitedSubjArr[1], ")"),
			"subject":      splitedSubjArr[0],
			"time":         strings.TrimSpace(item.Children().Eq(1).Text()),
			"attend":       strings.TrimSpace(item.Children().Eq(2).Text()),
			"late":         strings.TrimSpace(item.Children().Eq(3).Text()),
			"absence":      strings.TrimSpace(item.Children().Eq(4).Text()),
			"approved":     strings.TrimSpace(item.Children().Eq(5).Text()),
			"menstrual":    strings.TrimSpace(item.Children().Eq(6).Text()),
			"early":        strings.TrimSpace(item.Children().Eq(7).Text()),
		})
	})

	return gin.H{
		"attendance": attendanceData,
	}
}
