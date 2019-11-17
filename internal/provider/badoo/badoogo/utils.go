package badoogo

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type pageDownloader struct {
	FileName string
}

func (badoo *BadooRequester) TakeScreenshot(fileName string) error {
	var raw []byte
	if err := badoo.run(badoo.wrap("take screenshot", chromedp.Screenshot(`#page > div.page__wrap`, &raw))); err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, raw, 0644)
}

func (badoo *BadooRequester) TracePage(fileName string) error {
	return badoo.run(&pageDownloader{FileName: fileName})
}

func (badoo *BadooRequester) SetDebug(enabled bool) {
	badoo.debugMode = enabled
}

func (pd *pageDownloader) Do(ctx context.Context) error {
	node, err := dom.GetDocument().Do(ctx)
	if err != nil {
		return err
	}

	page, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
	if err == nil {
		return ioutil.WriteFile(pd.FileName, []byte(page), 0644)
	}

	return err
}

func (badoo *BadooRequester) run(actions ...chromedp.Action) error {
	return chromedp.Run(badoo.ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			timeoutContext, cancel := context.WithTimeout(ctx, opTimeout)
			defer cancel()
			var tasks chromedp.Tasks
			tasks = actions
			return tasks.Do(timeoutContext)
		}))
}

func (badoo *BadooRequester) wrap(text string, action chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if badoo.debugMode {
			badoo.log.Debugf("Running stage: %s", text)
		}
		err := chromedp.Run(ctx, action)
		if err != nil {
			err = fmt.Errorf("badoo: '%s': %+v", text, err)
			badoo.log.Error(err)
		}

		return err
	})
}
