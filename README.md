# headless-chrome

```go
package services

import (
	"context"
	"os"
	"github.com/pkg/errors"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func Screenshot(html string, buf *[]byte) error {
	allocCtx, cancel := chromedp.NewRemoteAllocator(
		context.Background(),
		os.GetEnv("CDP_URL", "ws://127.0.0.1:9222"),
	)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx, screenshotTasks(html, "body > div", buf)); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func screenshotTasks(html, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate("about:blank"),
		setScreenshotContent(html),
		emulation.SetDeviceMetricsOverride(0, 0, 1.0, false),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}

func setScreenshotContent(content string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		t, err := page.GetResourceTree().Do(ctx)
		if err != nil {
			return err
		}
		ch := cdpListenLoaded(ctx)
		if err = page.SetDocumentContent(t.Frame.ID, content).Do(ctx); err != nil {
			return err
		}
		<-ch
		return nil
	})
}

func cdpListenLoaded(ctx context.Context) chan bool {
	ch := make(chan bool)
	ctx, cancel := context.WithCancel(ctx)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		_, ok := ev.(*page.EventLoadEventFired)
		if ok {
			cancel()
			close(ch)
		}
	})
	return ch
}
```
