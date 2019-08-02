package users

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/sukso96100/skhus-backend/consts"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Userid string `form:"userid" json:"user" xml:"user"  binding:"required"`
	Userpw string `form:"userpw" json:"user" xml:"user"  binding:"required"`
}

func Login(c *gin.Context) {
	var loginData LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.String(http.StatusBadRequest,
			`Wrong login data form.
		올바르지 않은 로그인 데이터 양식입니다.`)
		return
	}

	if utf8.RuneCountInString(loginData.Userid) < 1 || utf8.RuneCountInString(loginData.Userpw) < 8 {
		c.String(http.StatusBadRequest,
			`ID or PW is empty. Or PW is shorter then 8 digits.
		If your using password with less then 8 digits, please change it at forest.skhu.ac.kr
		학번 또는 비밀번호가 비어있거나 비밀번호가 8자리 미만 입니다.
		8자리 미만 비밀번호 사용 시, forest.skhu.ac.kr 에서 변경 후 사용해 주세요.`)
	}

	// Options for custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE))

	// Create contexts
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	forestCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	samCtx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	credentialOld := make(chan string)
	credentialNew := make(chan string)
	credentialNewToken := make(chan string)
	loginError := make(chan string)

	// chromedp.Run(forestCtx, loginOnForest(loginData.Userid, loginData.Userpw, credentialOld, loginError))
	// chromedp.Run(samCtx, loginOnSam(loginData.Userid, loginData.Userpw, credentialNew, credentialNewToken, loginError))
	go loginOnForest(forestCtx, &loginData, credentialOld, loginError)
	go loginOnSam(samCtx, &loginData, credentialNew, credentialNewToken, loginError)

	// for {
	// 	select {}
	// }
}

func loginOnForest(ctx context.Context, loginData *LoginData,
	credentialOld chan string, loginError chan string) chromedp.Tasks {

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*page.EventFrameNavigated); ok {
			fmt.Println("closing alert:", ev.Message)
			go func() {
				if err := chromedp.Run(ctx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					panic(err)
				}
			}()
		}
	})

	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(fmt.Sprintf("%s}/Gate/UniLogin.aspx", consts.ForestUrl)),
		chromedp.WaitReady(`#txtID`),
		chromedp.SendKeys(`#txtID`, loginData.Userid),
		chromedp.SendKeys(`#txtPW`, loginData.Userpw),
	})

}

func loginOnSam(ctx context.Context, loginData *LoginData,
	credentialNew chan string, credentialNewToken chan string,
	loginError chan string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(consts.SkhuSamUrl),
	}
}
