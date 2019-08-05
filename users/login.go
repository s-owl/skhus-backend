package users

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/sukso96100/skhus-backend/consts"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/gin-gonic/gin"
)

type LoginData struct {
	Userid string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
	Userpw string `form:"userpw" json:"userpw" xml:"userpw"  binding:"required"`
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

	credentialOldChan := make(chan string)
	credentialNewChan := make(chan string)
	credentialNewTokenChan := make(chan string)
	loginErrorChan := make(chan string)

	var credentialOld, credentialNew, credentialNewToken string

	// chromedp.Run(forestCtx, loginOnForest(loginData.Userid, loginData.Userpw, credentialOld, loginError))
	// chromedp.Run(samCtx, loginOnSam(loginData.Userid, loginData.Userpw, credentialNew, credentialNewToken, loginError))
	go loginOnForest(forestCtx, &loginData, credentialOldChan, loginErrorChan)
	go loginOnSam(samCtx, &loginData, credentialNewChan, credentialNewTokenChan, loginErrorChan)

	for {
		select {
		case errorMsg := <-loginErrorChan:
			c.String(http.StatusUnauthorized, errorMsg)
			return
		case credentialOld = <-credentialOldChan:
			if credentialNew != "" && credentialNewToken != "" {
				break
			}
		case credentialNew = <-credentialNewChan:
			if credentialOld != "" && credentialNewToken != "" {
				break
			}
		case credentialNewToken = <-credentialNewTokenChan:
			if credentialOld != "" && credentialNew != "" {
				break
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"credential-old":       credentialOld,
		"credential-new":       credentialNew,
		"credential-new-token": credentialNewToken,
	})
}

func loginOnForest(ctx context.Context, loginData *LoginData,
	credentialOld chan string, loginError chan string) {
	loginPageURL := fmt.Sprintf("%s/Gate/UniLogin.aspx", consts.ForestURL)
	agreementPageURL := fmt.Sprintf("%s/Gate/CORE/P/CORP02P.aspx", consts.ForestURL)
	mainPageURL := fmt.Sprintf("%s/Gate/UniMyMain.aspx", consts.ForestURL)
	triedLogin := false

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventFrameNavigated); ok {
			targets, _ := chromedp.Targets(ctx)
			currentURL := targets[0].URL
			fmt.Println("Page URL", currentURL)
			switch currentURL {
			case loginPageURL:
				if triedLogin {
					errorMsg :=
						`Login Failed: Can't log in to forest.skhu.ac.kr, Check ID and PW again.
						로그인 실패: (forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
					loginError <- errorMsg
					break
				}
			case agreementPageURL:
				errorMsg :=
					`Please complete the privacy policy agreement on forest.skhu.ac.kr
					forest.skhu.ac.kr 에서 개인정보 제공 동의를 먼저 완료해 주세요.`
				loginError <- errorMsg
				break
			case mainPageURL:
				chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
					cookies, err := network.GetAllCookies().Do(ctx)
					if err != nil {
						return err
					}

					var buf bytes.Buffer
					for _, cookie := range cookies {
						buf.WriteString(fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value))
					}

					credentialOld <- buf.String()
					return nil
				}))
			}
		}
	})

	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(loginPageURL),
		chromedp.WaitReady(`#txtID`),
		chromedp.SetValue(`#txtID`, loginData.Userid, chromedp.ByID),
		chromedp.SetValue(`#txtPW`, loginData.Userpw, chromedp.ByID),
		chromedp.SendKeys(`#txtPW`, kb.Enter, chromedp.ByID),
	})
	triedLogin = true
}

func loginOnSam(ctx context.Context, loginData *LoginData,
	credentialNew chan string, credentialNewToken chan string,
	loginError chan string) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventFrameNavigated); ok {
			targets, _ := chromedp.Targets(ctx)
			currentURL := targets[0].URL
			fmt.Println("Page URL", currentURL)
			switch {
			case currentURL == consts.SkhuCasURL:
				chromedp.Run(ctx, chromedp.Tasks{
					chromedp.WaitReady(`#login-username`),
					chromedp.SetValue(`#login-username`, loginData.Userid, chromedp.ByID),
					chromedp.SetValue(`#login-password`, loginData.Userpw, chromedp.ByID),
					chromedp.SendKeys(`#login-password`, kb.Enter, chromedp.ByID),
					chromedp.WaitVisible(`body.ng-scope.modal-open`, chromedp.ByQuery),
				})
				errorMsg :=
					`Login Failed: Can't log in to sam.skhu.ac.kr, Check ID and PW again.
					로그인 실패: sam.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
				loginError <- errorMsg
			case strings.HasPrefix(currentURL, consts.SkhuSamURL):
				var tmpToken string
				chromedp.Run(ctx, chromedp.Tasks{
					chromedp.ActionFunc(func(ctx context.Context) error {
						cookies, err := network.GetAllCookies().Do(ctx)
						if err != nil {
							return err
						}

						var buf bytes.Buffer
						for _, cookie := range cookies {
							buf.WriteString(fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value))
						}

						credentialNew <- buf.String()
						return nil
					}),
					chromedp.Evaluate(`document.body.getAttribute("ncg-request-verification-token")`, &tmpToken),
				})
				credentialNewToken <- tmpToken
			}
		}
	})
	chromedp.Run(ctx, chromedp.Navigate(consts.SkhuSamURL))
}
