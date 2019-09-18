package browser

import (
	"log"
	"sync"
	"time"
	"context"

	"github.com/s-owl/skhus-backend/consts"

	"github.com/chromedp/chromedp"
)

// chromedp의 ExecAllocator를 singletone으로 관리하는 구조체(접근제어를 위해 소문자)
type browser struct {
	allocCtx context.Context
	cancelCtx context.CancelFunc
	*sync.RWMutex
}

func (brow *browser) isOk() bool {
	return brow.allocCtx != nil
}

func (brow *browser) init() {
	brow.Lock()
	defer brow.Unlock()
	if !brow.isOk() {
		// Options for custom user agent
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent(consts.UserAgentIE))
		// Create contexts
		brow.allocCtx, brow.cancelCtx =
			chromedp.NewExecAllocator(context.Background(), opts...)
		//특정 시간 후에 브라우저를 닫는다
		go func() {
			<-time.After(1*time.Minute)
			brow.close()
		}()
	}
}

func (brow *browser) close() {
	brow.Lock()
	defer brow.Unlock()
	brow.cancelCtx()
	brow.allocCtx = nil
}

var singletone *browser = &browser{
	RWMutex: &sync.RWMutex{},
}

// GetBrowser 브라우저를 가져온다.
func GetBrowser() *browser {
	// 초기화되지 않을 때 초기화
	if !singletone.isOk() {
		singletone.init()
	}
	return singletone
}

// NewContext 컨텍스트 생성
func (brow *browser)NewContext() (ctx context.Context, cf context.CancelFunc) {
	brow.RLock()
	// 최악의 경우의 수인 브라우저 닫히고 컨텍스트를 생성할 때
	if !brow.isOk() {
		brow.RUnlock()
		brow.init()
		brow.RLock()
	}
	ctx, cf = chromedp.NewContext(brow.allocCtx, chromedp.WithLogf(log.Printf))
	// 컨텍스트 종료시 뮤텍스 해제
	go func() {
		<-ctx.Done()
		brow.RUnlock()
	}()
	return
}
