package enroll

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

var targetURL = fmt.Sprintf("%s/GATE/SAM/LECTURE/S/SSGS09S.ASPX?&maincd=O&systemcd=S&seq=1", consts.ForestURL)

func GetSubjects(c *gin.Context) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", targetURL, nil)
	req.Header.Add("Cookie", c.MustGet("CredentialOld").(string))
	res, err := client.Do(req)
	if err != nil {
		c.String(http.StatusInternalServerError, consts.InternalError)
		return
	}
	defer res.Body.Close()

	c.JSON(http.StatusOK, extractData(tools.EucKrReaderToUtf8Reader(res.Body)))
}

type SubjectOption struct {
	Year      string `form:"year" json:"year" xml:"year"  binding:"required"`
	Semester  string `form:"semester" json:"semester" xml:"semester"  binding:"required"`
	Major     string `form:"major" json:"major" xml:"major"  binding:"required"`
	Professor string `form:"professor" json:"professor" xml:"professor"  binding:"required"`
}

func GetSubjectsWithOptions(c *gin.Context) {
	cookies := c.MustGet("CredentialOldCookies").([]*http.Cookie)
	var optionData SubjectOption
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
		chromedp.SetValue(`#ddlSosog`, optionData.Major, chromedp.ByQuery),
		chromedp.SetValue(`#txtPermNm`, optionData.Professor, chromedp.ByQuery),
		chromedp.Click(`#CSMenuButton1_List`, chromedp.ByQuery),
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
	list := []gin.H{}
	doc.Find("#dgList > tbody > tr").Each(func(i int, item *goquery.Selection) {
		if i > 0 && i < 13 {
			list = append(list, gin.H{
				"type":        item.Children().Eq(0).Text(),
				"grade":       item.Children().Eq(1).Text(),
				"code":        item.Children().Eq(2).Text(),
				"class":       item.Children().Eq(3).Text(),
				"subject":     item.Children().Eq(4).Text(),
				"score":       item.Children().Eq(5).Text(),
				"professor":   item.Children().Eq(6).Text(),
				"grade_limit": item.Children().Eq(7).Text(),
				"major_limit": item.Children().Eq(8).Text(),
				"time":        item.Children().Eq(9).Text(),
				"note":        item.Children().Eq(10).Text(),
				"available":   item.Children().Eq(11).Text(),
			})
		}
	})

	semesterOptions := []gin.H{}
	doc.Find("#ddlHaggi > option").Each(func(i int, item *goquery.Selection) {
		semesterOptions = append(semesterOptions, gin.H{
			"title": item.Text(),
			"value": item.AttrOr("value", ""),
		})
	})

	majorOptions := []gin.H{}
	doc.Find("#ddlSosog > option").Each(func(i int, item *goquery.Selection) {
		majorOptions = append(majorOptions, gin.H{
			"title": item.Text(),
			"value": item.AttrOr("value", ""),
		})
	})

	majorCurrent := doc.Find("#ddlSosog > option[selected=\"selected\"]")

	return gin.H{
		"list": list,
		"options": gin.H{
			"semester": semesterOptions,
			"major":    majorOptions,
			"major_current": gin.H{
				"title": majorCurrent.Text(),
				"value": majorCurrent.AttrOr("value", ""),
			},
		},
	}
}
