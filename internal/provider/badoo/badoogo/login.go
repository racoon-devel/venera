package badoogo

import (
	"fmt"
	"time"

	"github.com/chromedp/cdproto/emulation"

	"github.com/chromedp/chromedp"
)

func (badoo *BadooRequester) Login(email, password string) error {

	err := badoo.run(
		chromedp.EmulateViewport(1920, 1024),
		badoo.wrap("set geo", emulation.SetGeolocationOverride().WithLatitude(44.786568).WithLongitude(20.448921)),
		badoo.wrap("go to login page", chromedp.Navigate("https://badoo.com/ru/signin/?f=top")),
		badoo.wrap("wait email field", chromedp.WaitVisible(`//input[@name="email"]`)),
		badoo.wrap("wait login field", chromedp.WaitVisible(`//input[@name="password"]`)),
		badoo.wrap("wait", chromedp.Sleep(time.Second*3)),
		badoo.wrap("fill email", chromedp.SendKeys(`//input[@name="email"]`, email)),
		badoo.wrap("fill password", chromedp.SendKeys(`//input[@name="password"]`, password)),
		badoo.wrap("click login button", chromedp.Click(loginButtonPath)),
		badoo.wrap("wait", chromedp.Sleep(2*time.Second)),
		badoo.wrap("go to dating", chromedp.Navigate(`https://badoo.com/encounters`)),
		badoo.wrap("wait profile page", chromedp.WaitVisible(sidebarUserPath)),
	)

	if err == nil {
		badoo.status = statusAuthorized
	}

	return err
}

func (badoo *BadooRequester) Logout() error {
	if badoo.status == statusInited {
		return fmt.Errorf("invalid status: unauthorized")
	}

	badoo.status = statusInited

	return badoo.run(
		chromedp.Click(`#app_s > div > div > div > div.scroll__inner > div > div.sidebar__section.sidebar__section--info > div > div.sidebar-info__user > div.sidebar-info__signout`, chromedp.NodeVisible),
		chromedp.WaitVisible(`body > aside > section > div.ovl__frame-inner.js-ovl-content > div > div.ovl__actions > div > div.btn.js-signout-immediately`),
		chromedp.Click(`body > aside > section > div.ovl__frame-inner.js-ovl-content > div > div.ovl__actions > div > div.btn.js-signout-immediately`),
	)
}
