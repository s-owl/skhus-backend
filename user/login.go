package user

import (
	"log"
	"sync"
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

// LoginData 로그인 요청 데이터
type LoginData struct {
	Userid string `form:"userid" json:"userid" xml:"userid"  binding:"required"`
	Userpw string `form:"userpw" json:"userpw" xml:"userpw"  binding:"required"`
}

// LoginError 로그인 에러를 확인하기 위한 타입
type LoginError uint8

const (
	_ = iota
	WrongForm LoginError = iota
	ForestError
	ForestAgree
	SamError
)

// Error 에러 메세지를 출력
func (code LoginError) Error() string {
	var msg string
	switch code {
	case WrongForm:
		msg = `ID or PW is empty. Or PW is shorter then 8 digits.
			If you are using password with less then 8 digits, please change it at forest.skhu.ac.kr
			학번 또는 비밀번호가 비어있거나 비밀번호가 8자리 미만 입니다.
			8자리 미만 비밀번호 사용 시, forest.skhu.ac.kr 에서 변경 후 사용해 주세요.`
	case ForestError:
		msg = `Login Failed: Can't log in on forest.skhu.ac.kr, Check ID and PW again.
			로그인 실패: forest.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.`
	case ForestAgree:
		msg = `Login Failed: Please complete privacy policy agreement at forest.skhu.ac.kr
			로그인 실패: forest.skhu.ac.kr 에서 개인정보 제공 동의를 완료해 주세요.`
	case SamError:
		msg = `Login Failed: Can't log in on sam.skhu.ac.kr, Check ID and PW again.
			If your account only works on fores.skhu.ac.kr, Please contact Sungkonghoe University Electric Computing Center
			로그인 실패: sam.skhu.ac.kr 에 로그인 할 수 없습니다. 학번과 비밀번호를 다시 확인하세요.
			forest.skhu.ac.kr 에서만 정상 로그인이 가능한 경우, 성공회대학교 전자계산소에 연락하세요.`
	}
	return msg
}

// LoginResult 로그인 결과를 모으는 객체입니다.
type LoginResult struct {
	Credentials map[string]string
	Err LoginError
	errMutex *sync.Mutex
	TriedForest bool
	*sync.WaitGroup
}

// SetErr 에러를 뮤텍스를 걸은 후 쓴다.
// 이미 에러가 있을 떄 덮어쓰지 못하게 한다.
func (res *LoginResult) SetErr(err LoginError) {
	res.errMutex.Lock()
	if res.Err == 0 {
		res.Err = err
	}
	res.errMutex.Unlock()
}

// Login 요청을 받아서 처리하는 함수
func Login(c *gin.Context) {
	loginData := LoginData{}
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
	// 로그인 데이터의 길이 최소 길이 검증
	if utf8.RuneCountInString(loginData.Userid) < 1 || utf8.RuneCountInString(loginData.Userpw) < 8 {
		return nil, WrongForm
	}

	// Create contexts
	Browser := browser.GetBrowser()
	forestCtx, cancelForestCtx := Browser.NewContext()
	defer cancelForestCtx()
	samCtx, cancelSamCtx := Browser.NewContext()
	defer cancelSamCtx()

	loginResult := &LoginResult {
		Credentials: make(map[string]string),
		errMutex: new(sync.Mutex),
		WaitGroup: new(sync.WaitGroup),
	}

	loginResult.Add(2)
	go loginOnForest(forestCtx, loginData, loginResult)
	go loginOnSam(samCtx, loginData, loginResult)

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

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := loginResult.Credentials["credential-old"]; ok {
				return
			}
			if _, ok := ev.(*page.EventFrameStoppedLoading); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) == 0 {
					return
				}
				currentURL := targets[0].URL
				log.Printf("Page URL " + currentURL)
				switch currentURL {
				case loginPageURL:
					if loginResult.TriedForest {
						defer loginResult.Done()
						loginResult.SetErr(ForestError)
						break
					}
				case agreementPageURL:
					defer loginResult.Done()
					loginResult.SetErr(ForestAgree)
					break
				case mainPageURL:
					log.Printf("Logged in on forest")
					if _, ok = loginResult.Credentials["credential-old"]; !ok {
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
							log.Printf(result)
							loginResult.Credentials["credential-old"] = result
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
	loginResult.TriedForest = true
}

func loginOnSam(ctx context.Context, loginData LoginData,
	loginResult *LoginResult) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if _, ok := loginResult.Credentials["credential-new"]; ok {
				return
			}
			if _, ok := ev.(*page.EventFrameNavigated); ok {
				targets, _ := chromedp.Targets(ctx)
				if len(targets) == 0 {
					return
				}
				currentURL := targets[0].URL
				log.Printf("Page URL " + currentURL)
				switch {
				case strings.HasPrefix(currentURL, consts.SkhuCasURL):
					log.Printf("Logging in on Sam...")
					chromedp.Run(ctx, chromedp.Tasks{
						chromedp.SendKeys(`#login-username`, loginData.Userid),
						chromedp.SendKeys(`#login-password`, loginData.Userpw),
						chromedp.SendKeys(`login-password`, kb.Enter, chromedp.ByID),
					})
				case strings.HasPrefix(currentURL, consts.SkhuSamURL):
					log.Printf("Logged in on Sam")
					if _, ok := loginResult.Credentials["credential-new"]; !ok {
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
								log.Printf(result)


								loginResult.Credentials["credential-new"] = result
								if tokenOK {
									loginResult.Credentials["credential-new-token"] = tmpToken
								}
								return nil
							}),
						})
					}
				}
			} else if ev, ok := ev.(*dom.EventAttributeModified); ok {
				if ev.Name == "class" && ev.Value == "ng-scope modal-open" {
					defer loginResult.Done()
					loginResult.SetErr(SamError)
					return
				}
			}
		}()
	})
	chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(consts.SkhuSamURL),
	})
}
