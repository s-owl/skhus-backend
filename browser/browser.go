package browser

import (
	"log"
	"context"

	"github.com/s-owl/skhus-backend/consts"

	"github.com/chromedp/chromedp"
)

// chromedp의 ExecAllocator를 singletone으로 관리하는 구조체(접근제어를 위해 소문자)
type browser struct {
	allocCtx context.Context
	cancelCtx context.CancelFunc
}

var singletone *browser

func New() *browser {
	if singletone == nil {
		// Options for custom user agent
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent(consts.UserAgentIE))

		// Create contexts
		allocCtx, cancelCtx := chromedp.NewExecAllocator(context.Background(), opts...)
		singletone = &browser {
			allocCtx: allocCtx,
			cancelCtx: cancelCtx,
		}
	}
	return singletone
}

func (brow *browser)NewContext() (context.Context, context.CancelFunc) {
	return chromedp.NewContext(brow.allocCtx, chromedp.WithLogf(log.Printf))
}
