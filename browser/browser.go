package browser

import (
	"log"
	"context"

	"github.com/s-owl/skhus-backend/consts"

	"github.com/chromedp/chromedp"
)

// Browser chromedp의 ExecAllocator를 관리하는 구조체
type Browser struct {
	allocCtx context.Context
	cancelCtx context.CancelFunc
}

// NewBrowser context를 받아 Browser를 초기화한다.
func NewBrowser(ctx context.Context) *Browser {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(consts.UserAgentIE))

	// Create contexts
	allocCtx, cancelCtx := chromedp.NewExecAllocator(ctx, opts...)
	browser := &Browser {
		allocCtx: allocCtx,
		cancelCtx: cancelCtx,
	}
	return browser
}

// NewContext 브라우저에서 tab을 사용할 수 있는 context를 생성한다.
func (brow *Browser)NewContext() (context.Context, context.CancelFunc) {
	return chromedp.NewContext(brow.allocCtx, chromedp.WithLogf(log.Printf))
}

// Close Browser 객체의 context를 캔슬한다.
func (brow *Browser)Close() {
	brow.cancelCtx()
}
