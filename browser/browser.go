package browser

import (
	"log"
	"sync"
	"time"
	"context"

	"github.com/s-owl/skhus-backend/consts"

	"github.com/chromedp/chromedp"
)

const BROWSER_TIMEOUT = 1

// chromedp의 ExecAllocator를 singletone으로 관리하는 구조체(접근제어를 위해 소문자)
type browser struct {
	allocCtx  context.Context
	cancelCtx context.CancelFunc
	mutex     *sync.RWMutex
}

func (brow *browser) isOk() bool {
	return brow.allocCtx != nil
}

func (brow *browser) init() {
	// Options for custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE))
	// Create contexts
	brow.allocCtx, brow.cancelCtx =
		chromedp.NewExecAllocator(context.Background(), opts...)
	//특정 시간 후에 브라우저를 닫는다
	go func() {
		<-time.After(BROWSER_TIMEOUT*time.Minute)
		brow.reopen()
	}()
}

func (brow *browser) reopen() {
	brow.mutex.Lock()
	defer brow.mutex.Unlock()
	brow.cancelCtx()
	brow.init()
}

var singletone *browser = &browser{
	nil, nil, &sync.RWMutex{},
}

// GetBrowser 브라우저를 가져온다.
func GetBrowser() *browser {
	// 초기화되지 않을 때 초기화
	if !singletone.isOk() {
		singletone.mutex.Lock()
		singletone.init()
		singletone.mutex.Unlock()
	}
	return singletone
}

// NewContext 컨텍스트 생성
func (brow *browser)NewContext() (ctx context.Context, cf context.CancelFunc) {
	brow.mutex.RLock()
	ctx, cf = chromedp.NewContext(brow.allocCtx, chromedp.WithLogf(log.Printf))
	// 컨텍스트 종료시 뮤텍스 해제
	go func() {
		<-ctx.Done()
		brow.mutex.RUnlock()
	}()
	return
}
