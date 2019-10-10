package user

import (
	"fmt"
	"log"
	"bytes"
	"context"
	"strings"
	"net/http"
	"unicode/utf8"

	"github.com/s-owl/skhus-backend/consts"
	"github.com/s-owl/skhus-backend/browser"

	"github.com/gin-gonic/gin"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/network"
)

// loginError 로그인 에러를 확인하기 위한 타입
type loginError uint8

const (
	// FailedParsing gin에서 파싱 실패
	FailedParsing loginError = iota + 1
	// WrongForm 로그인 조건 불충분
	WrongForm
	// ForestError 포레스트 로그인 실패
	ForestError
	// ForestAgree 포레스트 개인정보 미동의
	ForestAgree
	// SamError SAM 로그인 실패
	SamError
)

// Error 에러 메세지를 출력
func (code loginError) Error() string {
	var msg string
	switch code {
	case FailedParsing:
		msg = `Wrong login data form.
올바르지 않은 로그인 데이터 양식입니다.
`
	case WrongForm:
		msg = `ID or PW is empty. Or PW is shorter then 8 digits.
If you are using password with less then 8 digits, please change it at forest.skhu.ac.kr
학번 또는 비밀번호가 비어있거나 비밀번호가 8자리 미만 입니다.
8자리 미만 비밀번호 사용 시, forest.skhu.ac.kr 에서 변경 후 사용해 주세요.
`
	case ForestError:
		msg = `Login Failed: Can't log in on forest.skhu.ac.kr, Check ID and PW again.
로그인 실패: forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.
`
	case ForestAgree:
		msg = `Login Failed: Please complete privacy policy agreement at forest.skhu.ac.kr
로그인 실패: forest.skhu.ac.kr 에서 개인정보 제공 동의를 완료해 주세요.
`
	case SamError:
		msg = `Login Failed: Can't log in on sam.skhu.ac.kr, Check ID and PW again.
If your account only works on fores.skhu.ac.kr, Please contact Sungkonghoe University Electric Computing Center
로그인 실패: sam.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.
forest.skhu.ac.kr 에서만 정상 로그인이 가능한 경우, 성공회대학교 전자계산소에 연락하세요.
`
	}
	return msg
}

// loginData 로그인 요청 데이터
type loginData struct {
	Userid string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
	Userpw string `form:"userpw" json:"userpw" xml:"userpw"  binding:"required"`
}

// getLoginData gin에서 loginData를 추출하고 실패시 에러 출력
func getLoginData(c *gin.Context) *loginData {
	data := &loginData{}
	// gin 컨텍스트에서 데이터 파싱
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%s\n%s", FailedParsing.Error(), err.Error()),
		})
		return nil
	}
	// 로그인 데이터의 길이 최소 길이 검증
	if utf8.RuneCountInString(data.Userid) < 1 || utf8.RuneCountInString(data.Userpw) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("%s", WrongForm.Error()),
		})
		return nil
	}
	return data
}

// 로그인 결과값
type loginResult interface {
	getErr() string
}

func response(c *gin.Context, res loginResult) {
	code := http.StatusOK
	if res.getErr() != "" {
		code = http.StatusUnauthorized
	}
	c.JSON(code, res)
}

type loginForestResult struct {
	Credential string `json:"credential"`
	Err        string `json:"error"`
}

func (res loginForestResult)getErr() string {
	return res.Err
}

func loginOnForest(ctx context.Context, userData *loginData) loginForestResult {
	loginPageURL := consts.ForestURL + "/Gate/UniLogin.aspx"
	agreementPageURL := consts.ForestURL + "/Gate/CORE/P/CORP02P.aspx"
	mainPageURL := consts.ForestURL + "/Gate/UniMyMain.aspx"

	result := make(chan loginForestResult)
	loginTried := false
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if result == nil {
				return
			} else if _, ok := ev.(*page.EventFrameStoppedLoading); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) == 0 {
					return
				}
				currentURL := targets[0].URL
				log.Printf("Page URL " + currentURL)
				switch currentURL {
				case loginPageURL:
					if loginTried {
						result <- loginForestResult {
							Err: ForestError.Error(),
						}
					}
				case agreementPageURL:
						result <- loginForestResult {
							Err: ForestAgree.Error(),
						}
				case mainPageURL:
					log.Printf("Logged in on forest")
					go chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
						cookies, err := network.GetAllCookies().Do(ctx)
						if err != nil {
							return err
						}

						var buf bytes.Buffer
						for _, cookie := range cookies {
							buf.WriteString(cookie.Name + "=" + cookie.Value + ";")
						}
						credential := buf.String()
						log.Printf(credential)
						result <- loginForestResult {
							Credential: credential,
						}
						return nil
					}))
				}
			}
		}()
	})

	go chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(loginPageURL),
		chromedp.WaitReady(`txtID`, chromedp.ByID),
		chromedp.SetValue(`txtID`, userData.Userid, chromedp.ByID),
		chromedp.SetValue(`txtPW`, userData.Userpw, chromedp.ByID),
		chromedp.SendKeys(`txtPW`, kb.Enter, chromedp.ByID),
		chromedp.ActionFunc(func(ctx context.Context) error {
			loginTried = true
			return nil
		}),
	})

	res := <-result
	result = nil
	return res
}

// LoginForest 포레스트 로그인
func LoginForest(c *gin.Context) {
	userData := getLoginData(c)
	if userData == nil {
		return
	}

	// browser 초기화
	Browser := browser.NewBrowser(c)
	defer Browser.Close()
	tab, cf := Browser.NewContext()
	defer cf()
	// 결과 전송
	response(c, loginOnForest(tab, userData))
}

type loginSamResult struct {
	Credential string `json:"credential"`
	Token      string `json:"token"`
	Err        string `json:"error"`
}

func (res loginSamResult)getErr() string {
	return res.Err
}

func loginOnSam(ctx context.Context, userData *loginData) loginSamResult {
	result := make(chan loginSamResult)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if result == nil {
				return
			} else if _, ok := ev.(*page.EventFrameNavigated); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) == 0 {
					return
				}
				currentURL := targets[0].URL
				log.Printf("Page URL " + currentURL)
				switch {
				case strings.HasPrefix(currentURL, consts.SkhuCasURL):
					log.Printf("Logging in on Sam...")
					go chromedp.Run(ctx, chromedp.Tasks{
						chromedp.SendKeys(`#login-username`, userData.Userid),
						chromedp.SendKeys(`#login-password`, userData.Userpw),
						chromedp.SendKeys(`login-password`, kb.Enter, chromedp.ByID),
					})
				case strings.HasPrefix(currentURL, consts.SkhuSamURL):
					log.Printf("Logged in on Sam")
					var tmpToken string
					var tokenOK bool
					go chromedp.Run(ctx, chromedp.Tasks{
						chromedp.AttributeValue(`body`, `ncg-request-verification-token`, &tmpToken, &tokenOK, chromedp.ByQuery),
						chromedp.ActionFunc(func(ctx context.Context) error {
							cookies, err := network.GetAllCookies().Do(ctx)
							if err != nil {
								return err
							}

							var buf bytes.Buffer
							for _, cookie := range cookies {
								buf.WriteString(cookie.Name + "=" + cookie.Value + ";")
							}

							credential := buf.String()
							log.Printf(credential)

							token := tmpToken
							result <- loginSamResult {
								Credential: credential,
								Token: token,
							}
							return nil
						}),
					})
				}
			} else if ev, ok := ev.(*dom.EventAttributeModified); ok {
				if ev.Name == "class" && ev.Value == "ng-scope modal-open" {
					result <- loginSamResult {
						Err: SamError.Error(),
					}
					return
				}
			}
		}()
	})
	go chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(consts.SkhuSamURL),
	})

	res := <-result
	result = nil
	return res
}

// LoginSam Sam 로그인
func LoginSam(c *gin.Context) {
	userData := getLoginData(c)
	if userData == nil {
		return
	}

	// Browser 초기화
	Browser := browser.NewBrowser(c)
	defer Browser.Close()
	tab, cf := Browser.NewContext()
	defer cf()
	// 결과 전송
	response(c, loginOnSam(tab, userData))
}

type totalResult struct {
	OldCredential      string `json:"credential"`
	NewCredential      string `json:"new-credential"`
	NewCredentialToken string `json:"new-credential-token"`
}

// Login 기존 로그인
func Login(c *gin.Context) {
	userData := &loginData{}
	// gin 컨텍스트에서 데이터 파싱
	if err := c.ShouldBindJSON(userData); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("%s\n%s", FailedParsing.Error(), err.Error()))
		return
	}
	// 로그인 데이터의 길이 최소 길이 검증
	if utf8.RuneCountInString(userData.Userid) < 1 || utf8.RuneCountInString(userData.Userpw) < 8 {
		c.String(http.StatusBadRequest, fmt.Sprintf("%s", WrongForm.Error()))
		return
	}

	// Browser 초기화
	Browser := browser.NewBrowser(c)
	defer Browser.Close()

	// 동시 처리를 위한 채널 초기화
	forestResult := make(chan loginForestResult)
	samResult := make(chan loginSamResult)
	// 동시 시작
	go func() {
		tab, cf := Browser.NewContext()
		defer cf()
		forestResult <- loginOnForest(tab, userData)
	}()
	go func() {
		tab, cf := Browser.NewContext()
		defer cf()
		samResult <- loginOnSam(tab, userData)
	}()

	// 결과 확인
	forest := <-forestResult
	sam := <-samResult
	// error 메세지 우선 순위는 forest가 우선
	if forest.Err != "" || sam.Err != "" {
		if forest.Err == "" {
			c.String(http.StatusUnauthorized, sam.Err)
		} else {
			c.String(http.StatusUnauthorized, forest.Err)
		}
		return
	}
	c.JSON(http.StatusOK, totalResult{
		forest.Credential,
		sam.Credential,
		sam.Token,
	})
}
