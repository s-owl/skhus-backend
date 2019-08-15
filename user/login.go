package user

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"errors"
	"sync/atomic"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/s-owl/skhus-backend/consts"

	"github.com/chromedp/cdproto/dom"
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

type LoginResult interface {
	Err() error
}

type OldLoginResult struct {
	CredentialOld	string
	err	error
}

func (res *OldLoginResult) Err() error {
	return res.err
}

type NewLoginResult struct {
	CredentialNew	string
	CredentialNewToken	string
	err error
}

func (res *NewLoginResult) Err() error {
	return res.err
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
		return
	}

	// Options for custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE))

	// Create contexts
	allocCtx, cancelCtx := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelCtx()
	forestCtx, cancelForestCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	samCtx, cancelSamCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	loginResultChan := make(chan LoginResult, 2)

	var oldLoginResult *OldLoginResult
	var newLoginResult *NewLoginResult

	var runner int32 = 0
	go loginOnForest(forestCtx, &loginData, loginResultChan, &runner)
	go loginOnSam(samCtx, &loginData, loginResultChan, &runner)

	for {
		res := <-loginResultChan
		if err := res.Err(); err != nil {
			c.String(http.StatusUnauthorized, err.Error())
			cancelSamCtx()
			cancelForestCtx()
			return
		}
		switch r := res.(type) {
			case *OldLoginResult:
			oldLoginResult = r
			fmt.Println("OldLogin Complete")
			cancelForestCtx()
			case *NewLoginResult:
			newLoginResult = r
			fmt.Println("NewLogin Complete")
			cancelSamCtx()
		}
		if oldLoginResult != nil && newLoginResult != nil {
			c.JSON(http.StatusOK, gin.H{
				"credential-old":       (*oldLoginResult).CredentialOld,
				"credential-new":       (*newLoginResult).CredentialNew,
				"credential-new-token": (*newLoginResult).CredentialNewToken,
			})
			return
		}
	}
}

func startLogin(runner *int32) {
	atomic.AddInt32(runner, 1)
}

func completeLogin(runner *int32, ch chan LoginResult) {
	current := atomic.AddInt32(runner, -1)
	if current == 0 {
		close(ch)
	}
}

func errorLogin(runner *int32, ch chan LoginResult) {
	atomic.StoreInt32(runner, 0)
	close(ch)
}

func loginOnForest(ctx context.Context, loginData *LoginData,
	loginResult chan LoginResult, runner *int32) {
	startLogin(runner)
	loginPageURL := consts.ForestURL + "/Gate/UniLogin.aspx"
	agreementPageURL :=  consts.ForestURL + "/Gate/CORE/P/CORP02P.aspx"
	mainPageURL := consts.ForestURL + "/Gate/UniMyMain.aspx"
	triedLogin := false
	isCredentialSent := false

	errorResult := func(errorMsg string) LoginResult {
		return &OldLoginResult{
			CredentialOld: "",
			err: errors.New(errorMsg),
		}
	}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := ev.(*page.EventFrameStoppedLoading); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) == 0 {
					return
				}
				currentURL := targets[0].URL
				fmt.Println("Page URL", currentURL)
				switch currentURL {
				case loginPageURL:
					if triedLogin {
						errorMsg :=
							`Login Failed: Can't log in to forest.skhu.ac.kr, Check ID and PW again.
							로그인 실패: (forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
						loginResult <- errorResult(errorMsg)
						errorLogin(runner, loginResult)
						break
					}
				case agreementPageURL:
					errorMsg :=
						`Please complete the privacy policy agreement on forest.skhu.ac.kr
						forest.skhu.ac.kr 에서 개인정보 제공 동의를 먼저 완료해 주세요.`
						loginResult <- errorResult(errorMsg)
						errorLogin(runner, loginResult)
					break
				case mainPageURL:
					fmt.Println("Logged in on forest")
					if !isCredentialSent {

						chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
							cookies, err := network.GetAllCookies().Do(ctx)
							if err != nil {
								return err
							}

							var buf bytes.Buffer
							for _, cookie := range cookies {
								buf.WriteString(fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value))
							}
							result := buf.String()
							fmt.Println(result)

							loginResult <- &OldLoginResult{
								CredentialOld: result,
								err: nil,
							}
							completeLogin(runner, loginResult)
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

func loginOnSam(ctx context.Context, loginData *LoginData,
	loginResult chan LoginResult, runner *int32) {
	startLogin(runner)
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

								var buf bytes.Buffer
								for _, cookie := range cookies {
									buf.WriteString(fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value))
								}

								result := buf.String()
								fmt.Println(result)

								login := &NewLoginResult{
									CredentialNew: result,
									err: nil,
								}
								if tokenOK {
									 login.CredentialNewToken = tmpToken
								}
								loginResult <- login
								completeLogin(runner, loginResult)
								isCredentialSent = true
								return nil
							}),
						})
					}
				}
			} else if ev, ok := ev.(*dom.EventAttributeModified); ok {
				if ev.Name == "class" && ev.Value == "ng-scope modal-open" {
					errorMsg :=
						`Login Failed: Can't log in to sam.skhu.ac.kr, Check ID and PW again.
						로그인 실패: sam.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
					loginResult <- &NewLoginResult{
						err: errors.New(errorMsg),
					}
					errorLogin(runner, loginResult)
				}
			}
		}()
	})
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(consts.SkhuSamURL),
	})
}
