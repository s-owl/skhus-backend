package user

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/tools"
)

var targetURL = fmt.Sprintf("%s/Gate/UniMainStudent.aspx", consts.ForestURL)

func GetCurrentAttendance(c *gin.Context) {
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

	// Options for custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE),
		// chromedp.Flag("headless", false)
	)

	// Create contexts
	allocCtx, cancelAllocCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAllocCtx()
	ctx, cancelCtx := chromedp.NewContext(allocCtx)
	defer cancelCtx()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := ev.(*page.EventFrameStoppedLoading); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) > 0 {
					currentURL := targets[0].URL
					fmt.Println("Page URL", currentURL)
				}
			}
		}()
	})
	var content string
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(consts.ForestURL),
		chromedp.ActionFunc(func(context context.Context) error {
			network.Enable().Do(context)

			// Set Cokies
			for _, item := range cookies {
				cookieParam := network.SetCookie(item.Name, item.Value)
				cookieParam.URL = targetURL
				ok, err := cookieParam.Do(context)
				if ok {
					fmt.Println("Cookie Set")
				} else if err != nil {
					fmt.Println(err)
				}
			}

			// Block CoreSecurity.js
			network.SetBlockedURLS(
				[]string{
					consts.CoreSecurity,
				}).Do(context)
			return nil
		}),
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(`txtYy`, chromedp.ByID),
		chromedp.SetValue(`#txtYy`, optionData.Year, chromedp.ByQuery),
		chromedp.SetValue(`#ddlHaggi`, optionData.Semester, chromedp.ByQuery),
		chromedp.Click(`#btnList`, chromedp.ByQuery),
		chromedp.Sleep(300 * time.Millisecond),
		chromedp.InnerHTML(`body`, &content, chromedp.ByQuery),
	})
	fmt.Println(content)
	c.JSON(http.StatusOK, extractData(strings.NewReader(content)))
	return
}

func extractData(body io.Reader) map[string]interface{} {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		log.Fatal(err)
	}
	attendanceData := []gin.H{}
	doc.Find("#gvList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		if i > 0 {
			subjectStr := strings.TrimSpace(item.Children().Eq(0).Text())
			splitedSubjArr := strings.Split(subjectStr, "(")
			attendanceData = append(attendanceData, gin.H{
				"subject_code": strings.Trim(splitedSubjArr[1], ")"),
				"subject":      strings.TrimSpace(splitedSubjArr[0]),
				"time":         strings.TrimSpace(item.Children().Eq(1).Text()),
				"attend":       strings.TrimSpace(item.Children().Eq(2).Text()),
				"late":         strings.TrimSpace(item.Children().Eq(3).Text()),
				"absence":      strings.TrimSpace(item.Children().Eq(4).Text()),
				"approved":     strings.TrimSpace(item.Children().Eq(5).Text()),
				"menstrual":    strings.TrimSpace(item.Children().Eq(6).Text()),
				"early":        strings.TrimSpace(item.Children().Eq(7).Text()),
			})
		}
	})

	return gin.H{
		"attendance": attendanceData,
	}
}