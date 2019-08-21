package user

import (
	"fmt"
	"sync"
	"bytes"
	"context"
	"strings"
	"net/http"

	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/browser"

	"github.com/gin-gonic/gin"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/network"
)

type LoginData struct {
	Userid string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
	Userpw string `form:"userpw" json:"userpw" xml:"userpw"  binding:"required"`
}

type LoginError uint8

const (
	_ = iota
	WrongForm LoginError = iota
	ForestError
	ForestAgree
	SamError
)

func (code LoginError) Error() string {
	var msg string
	switch code {
	case WrongForm:
		msg = `ID or PW is empty. Or PW is shorter then 8 digits.
			If your using password with less then 8 digits, please change it at forest.skhu.ac.kr
			학번 또는 비밀번호가 비어있거나 비밀번호가 8자리 미만 입니다.
			8자리 미만 비밀번호 사용 시, forest.skhu.ac.kr 에서 변경 후 사용해 주세요.`
	case ForestError:
		msg = `Login Failed: Can't log in to forest.skhu.ac.kr, Check ID and PW again.
			로그인 실패: (forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
	case ForestAgree:
		msg = `Login Failed: Can't log in to forest.skhu.ac.kr, Check ID and PW again.
			로그인 실패: (forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
	case SamError:
		msg = `Login Failed: Can't log in to sam.skhu.ac.kr, Check ID and PW again.
			로그인 실패: sam.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
	}
	return msg
}

type LoginResult struct {
	Credentials map[string]string
	Err LoginError
	sync.WaitGroup
}

func Login(c *gin.Context) {

	var loginData LoginData
	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.String(http.StatusBadRequest,
			`Wrong login data form.
			올바르지 않은 로그인 데이터 양식입니다.`)
		return
	}

	if res, err := runLogin(loginData); err != 0 {
		c.String(http.StatusUnauthorized, err.Error())
	} else {
		c.JSON(http.StatusOK, res)
	}
	return
}

func runLogin(loginData LoginData) (map[string]string, LoginError) {
	// Create contexts
	brow := browser.New()
	forestCtx, cancelForestCtx := brow.NewContext()
	defer cancelForestCtx()
	samCtx, cancelSamCtx := brow.NewContext()
	defer cancelSamCtx()

	loginResult := LoginResult {
		Credentials: make(map[string]string),
	}

	loginResult.Add(2)
	go loginOnForest(forestCtx, loginData, &loginResult)
	go loginOnSam(samCtx, loginData, &loginResult)

	loginResult.Wait()
	if loginResult.Err != 0 {
		return nil, loginResult.Err
	}

	return loginResult.Credentials, 0
}

func loginOnForest(ctx context.Context, loginData LoginData,
	loginResult *LoginResult) {
	loginPageURL := consts.ForestURL + "/Gate/UniLogin.aspx"
	agreementPageURL := consts.ForestURL + "/Gate/CORE/P/CORP02P.aspx"
	mainPageURL := consts.ForestURL + "/Gate/UniMyMain.aspx"
	triedLogin := false
	isCredentialSent := false

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := ev.(*page.EventFrameStoppedLoading); ok {
				targets, _ := chromedp.Targets(ctx)
				currentURL := targets[0].URL
				fmt.Println("Page URL", currentURL)
				switch currentURL {
				case loginPageURL:
					if triedLogin {
						defer loginResult.Done()
						loginResult.Err = ForestError
						break
					}
				case agreementPageURL:
					defer loginResult.Done()
					loginResult.Err = ForestAgree
					break
				case mainPageURL:
					fmt.Println("Logged in on forest")
					if !isCredentialSent {
						chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
							cookies, err := network.GetAllCookies().Do(ctx)
							if err != nil {
								return err
							}

							defer loginResult.Done()

							var buf bytes.Buffer
							for _, cookie := range cookies {
								buf.WriteString(cookie.Name + "=" + cookie.Value + ";")
							}
							result := buf.String()
							fmt.Println(result)

							loginResult.Credentials["credential-old"] = result
							isCredentialSent = true
							return nil
						}))
					}
				}
			}
		}()
	})

	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(loginPageURL),
		chromedp.WaitReady(`txtID`, chromedp.ByID),
		chromedp.SetValue(`txtID`, loginData.Userid, chromedp.ByID),
		chromedp.SetValue(`txtPW`, loginData.Userpw, chromedp.ByID),
		chromedp.SendKeys(`txtPW`, kb.Enter, chromedp.ByID),
	})
	triedLogin = true
}

func loginOnSam(ctx context.Context, loginData LoginData,
	loginResult *LoginResult) {
	isCredentialSent := false
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := ev.(*page.EventFrameNavigated); ok {
				targets, _ := chromedp.Targets(ctx)
				currentURL := targets[0].URL
				fmt.Println("Page URL", currentURL)
				switch {
				case strings.HasPrefix(currentURL, consts.SkhuCasURL):
					fmt.Println("Logging in on Sam...")
					chromedp.Run(ctx, chromedp.Tasks{
						chromedp.SendKeys(`#login-username`, loginData.Userid),
						chromedp.SendKeys(`#login-password`, loginData.Userpw),
						chromedp.SendKeys(`login-password`, kb.Enter, chromedp.ByID),
					})
				case strings.HasPrefix(currentURL, consts.SkhuSamURL):
					fmt.Println("Logged in on Sam")
					if !isCredentialSent {
						var tmpToken string
						var tokenOK bool
						chromedp.Run(ctx, chromedp.Tasks{
							chromedp.AttributeValue(`body`, `ncg-request-verification-token`, &tmpToken, &tokenOK, chromedp.ByQuery),
							chromedp.ActionFunc(func(ctx context.Context) error {
								cookies, err := network.GetAllCookies().Do(ctx)
								if err != nil {
									return err
								}

								defer loginResult.Done()
								var buf bytes.Buffer
								for _, cookie := range cookies {
									buf.WriteString(cookie.Name + "=" + cookie.Value + ";")
								}

								result := buf.String()
								fmt.Println(result)

								loginResult.Credentials["credential-new"] = result
								if tokenOK {
									loginResult.Credentials["credential-new-token"] = result
								}
								isCredentialSent = true
								return nil
							}),
						})
					}
				}
			} else if ev, ok := ev.(*dom.EventAttributeModified); ok {
				if ev.Name == "class" && ev.Value == "ng-scope modal-open" {
					defer loginResult.Done()
					loginResult.Err = SamError
					return
				}
			}
		}()
	})
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(consts.SkhuSamURL),
	})
}
