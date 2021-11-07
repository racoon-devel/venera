package navigator

import (
	"io/ioutil"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/mxschmitt/playwright-go"
)

const defaultTimeout float64 = 1000.0 * 10

var env *playwright.Playwright

type Navigator struct {
	browser playwright.Browser
	page    playwright.Page

	log      *logging.Logger
	err      error
	batch    string
	dumpPath string
}

func Open(log *logging.Logger, url string) (n *Navigator, err error) {
	if env == nil {
		if env, err = playwright.Run(); err != nil {
			return
		}
	}

	b := &Navigator{log: log}
	if b.browser, err = env.Firefox.Launch(); err != nil {
		return
	}

	if b.page, err = b.browser.NewPage(); err != nil {
		return
	}

	log.Debugf("browser: opening '%s'...", url)

	if _, err = b.page.Goto(url); err != nil {
		return
	}

	b.page.SetDefaultTimeout(defaultTimeout)

	n = b
	return
}

func (n *Navigator) Batch(title string) *Navigator {
	n.batch = title
	return n
}

func (n *Navigator) Goto(url string) *Navigator {
	if n.err != nil {
		return n
	}
	n.log.Debugf("%s: navigating to '%s'...", n.batch, url)
	_, n.err = n.page.Goto(url)
	n.checkError("Goto")
	return n
}

func (n *Navigator) Error() error {
	return n.err
}

func (n *Navigator) ClearError() *Navigator {
	n.err = nil
	return n
}

func (n *Navigator) Type(selector, text string) *Navigator {
	if n.err != nil {
		return n
	}
	n.log.Debugf("%s: typing '%s' to '%s'...", n.batch, text, selector)

	_, n.err = n.page.WaitForSelector(selector)
	if n.err == nil {
		n.err = n.page.Fill(selector, text)
	}
	n.checkError("Type")
	return n
}

func (n *Navigator) Click(selector string) *Navigator {
	if n.err != nil {
		return n
	}
	n.log.Debugf("%s: clicking on '%s'...", n.batch, selector)
	_, n.err = n.page.WaitForSelector(selector)
	if n.err == nil {
		n.err = n.page.Click(selector)
	}
	n.checkError("Click")
	return n
}

func (n *Navigator) Fetch(selector string, output *string) *Navigator {
	if n.err != nil {
		return n
	}
	n.log.Debugf("%s: fetching '%s'...", n.batch, selector)
	_, n.err = n.page.WaitForSelector(selector)
	if n.err == nil {
		var s playwright.ElementHandle
		s, n.err = n.page.QuerySelector(selector)
		if n.err == nil {
			*output, n.err = s.InnerText()
		}
	}
	n.checkError("Click")
	return n
}

func (n *Navigator) Screenshot(fileName string) *Navigator {
	data, err := n.page.Screenshot()
	if err != nil {
		n.log.Warnf("Cannot get screenshot: %+v", err)
		return n
	}
	if err = ioutil.WriteFile(fileName, data, 0644); err != nil {
		n.log.Warnf("Cannot save screenshot: %+v", err)
	}
	return n
}

func (n *Navigator) TracePage(fileName string) *Navigator {
	content, err := n.page.Content()
	if err != nil {
		n.log.Warnf("Cannot fetch page content: %+v", err)
		return n
	}
	if err = ioutil.WriteFile(fileName, []byte(content), 0644); err != nil {
		n.log.Warnf("Cannot save page content: %+v", err)
	}
	return n
}

func (n *Navigator) Sleep(d time.Duration) *Navigator {
	if n.err != nil {
		return n
	}
	n.log.Debugf("%s: waiting for %+v", n.batch, d)
	<-time.After(d)
	return n
}
