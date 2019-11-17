package badoogo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	ActionPass = iota
	ActionLike
	ActionSkip
	ActionStop
)

const (
	presentDataTimeout = 100 * time.Millisecond
)

type PersonRater func(user *BadooUser) int

func (badoo *BadooRequester) Spawn() (*BadooRequester, error) {

	if badoo.status == statusInited {
		return nil, fmt.Errorf("invalid status: unauthorized")
	}

	b := newRequester(badoo.log)
	b.ctx, b.cancel = chromedp.NewContext(badoo.ctx)
	b.status = statusAuthorized
	b.debugMode = badoo.debugMode

	return b, b.run(
		chromedp.EmulateViewport(1920, 1024),
		b.wrap("go to badoo.com", chromedp.Navigate("https://badoo.com")),
		b.wrap("wait sidebar", chromedp.WaitVisible(sidebarUserPath)),
	)
}

func (badoo *BadooRequester) WalkAround(handler PersonRater) error {
	if err := badoo.goWalking(); err != nil {
		return err
	}

	walker, err := badoo.Spawn()
	if err != nil {
		return err
	}

	defer walker.Close()

	dir := make([]map[string]string, 0)
	err = badoo.run(
		badoo.wrap("wait profiles", chromedp.WaitVisible(`a.user-card__link`)),
		badoo.wrap("fetch profiles", chromedp.AttributesAll(`a.user-card__link`, &dir)),
		badoo.wrap("go to next", chromedp.Click(`a.pagination__nav--next`)),
	)

	if err != nil {
		return err
	}

	for _, attrs := range dir {
		url, ok := attrs["href"]
		if ok {
			user, err := walker.FetchProfile("https://badoo.com" + url)
			if err != nil {
				walker.log.Warningf("Extract profile '%s' failed: %+v", url, err)
				continue
			}

			switch action := handler(user); action {
			case ActionStop:
				return nil
			case ActionLike:
				walker.run(
					walker.wrap("like", chromedp.Click(likeButtonPath)),
					walker.wrap("sleep", chromedp.Sleep(presentDataTimeout)),
				)
			case ActionPass:
				walker.run(
					walker.wrap("pass", chromedp.Click(passButtonPath)),
					walker.wrap("sleep", chromedp.Sleep(presentDataTimeout)),
				)
			}
		}
	}

	return nil
}

func (badoo *BadooRequester) Fetch(handler PersonRater) error {
	if err := badoo.goDating(); err != nil {
		return err
	}

	user := &BadooUser{}
	age := ""
	err := badoo.run(
		badoo.wrap("extract profile name", chromedp.Text(profileNamePath, &user.Name)),
		badoo.wrap("extract age", chromedp.Text(profileAgePath, &age)),
		badoo.wrap("extract job", extractIfExists(profileJobPath, &user.Job)),
		badoo.wrap("extract about", extractIfExists(profileAboutPath, &user.About)),
		badoo.wrap("extract photos", extractPhotos(badoo, user)),
		badoo.wrap("wait", chromedp.Sleep(3*presentDataTimeout)),
		badoo.wrap("go to profile info", chromedp.Click(cardLinkPath, chromedp.NodeVisible)),
		badoo.wrap("wait", chromedp.Sleep(3*presentDataTimeout)),
		badoo.wrap("wait like button", chromedp.WaitVisible(likeButtonPath)),
		badoo.wrap("extract target", chromedp.Text(profileTargetPath, &user.Target)),
		badoo.wrap("extract profile info", extractIfExists(`div.personal-info`, &user.Info)),
		badoo.wrap("extract interests", extractIfExists(`div.pills`, &user.Interests)),
	)

	if err != nil {
		return err
	}

	val, _ := strconv.ParseInt(badoo.ageExpr.FindString(age), 10, 8)
	user.Age = int(val)

	result := handler(user)

	switch result {
	case ActionLike:
		err = badoo.run(
			badoo.wrap("like", chromedp.Click(likeButtonPath)),
			badoo.wrap("sleep", chromedp.Sleep(presentDataTimeout)),
		)
	case ActionPass:
		err = badoo.run(
			badoo.wrap("pass", chromedp.Click(passButtonPath)),
			badoo.wrap("sleep", chromedp.Sleep(presentDataTimeout)),
		)
	default:
		return fmt.Errorf("invalid return action")
	}

	return err
}

func (badoo *BadooRequester) FetchProfile(url string) (*BadooUser, error) {
	user := &BadooUser{}

	if badoo.status == statusInited {
		return nil, fmt.Errorf("invalid status: unauthorized")
	}

	age := ""

	err := badoo.run(
		badoo.wrap("open profile", chromedp.Navigate(url)),
		badoo.wrap("wait profile name", chromedp.WaitVisible(profileNamePath)),
		badoo.wrap("extract profile name", chromedp.Text(profileNamePath, &user.Name)),
		badoo.wrap("extract age", chromedp.Text(profileAgePath, &age)),
		badoo.wrap("extract job", extractIfExists(profileJobPath, &user.Job)),
		badoo.wrap("extract about", extractIfExists(`personal-info__about`, &user.About)),
		badoo.wrap("extract target", chromedp.Text(profileTargetPath, &user.Target)),
		badoo.wrap("extract profile info", extractIfExists(`div.personal-info`, &user.Info)),
		badoo.wrap("extract interests", extractIfExists(`div.pills`, &user.Interests)),
		badoo.wrap("click for photos", chromedp.Click(`span.js-profile-header-name`)),
		extractPhotos2(badoo, user),
	)

	val, _ := strconv.ParseInt(badoo.ageExpr.FindString(age), 10, 8)
	user.Age = int(val)

	matches := badoo.idExpr.FindStringSubmatch(url)
	if len(matches) >= 2 {
		user.URL = matches[0]
		user.ID = matches[1]
	}

	return user, err
}

func extractIfExists(selector string, result *string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		timeouted, cancel := context.WithTimeout(ctx, presentDataTimeout)
		defer cancel()
		chromedp.Run(timeouted, chromedp.Text(selector, result))
		return nil
	})
}

func extractPhotos(badoo *BadooRequester, user *BadooUser) chromedp.ActionFunc {
	user.Photos = make([]string, 0)
	return func(ctx context.Context) error {
		url := ""
		ok := false
		var prevN uint64

		for {
			total := ""
			current := ""

			err := chromedp.Run(ctx,
				badoo.wrap("extract total", chromedp.Text(`span.js-gallery-photo-total.js-gallery-total`, &total)),
				badoo.wrap("extract current", chromedp.Text(`span.js-gallery-photo-current`, &current)),
				badoo.wrap("extract photo url", chromedp.AttributeValue(`div.js-mm-photo`, "style", &url, &ok)),
			)

			if err != nil {
				return err
			}

			totalN, _ := strconv.ParseUint(total, 10, 16)
			currentN, _ := strconv.ParseUint(current, 10, 16)

			if currentN == prevN {
				badoo.wrap("go to next photo", chromedp.Click(`span.photo-gallery__link.photo-gallery__link--next.js-gallery-next`)).Do(ctx)
				continue
			}

			badoo.log.Debugf("photo %d / %d", currentN, totalN)

			if !ok {
				break
			}

			matched := badoo.urlExpr.FindStringSubmatch(url)
			if len(matched) > 1 {
				url = matched[1]
			} else {
				break
			}

			url = strings.Trim(url, `"`)

			user.Photos = append(user.Photos, url)

			if currentN == totalN {
				break
			}

			prevN = currentN

			badoo.wrap("go to next photo", chromedp.Click(`span.photo-gallery__link.photo-gallery__link--next.js-gallery-next`)).Do(ctx)
		}

		return nil
	}
}

func extractPhotos2(badoo *BadooRequester, user *BadooUser) chromedp.ActionFunc {
	user.Photos = make([]string, 0)
	return func(ctx context.Context) error {
		url := ""
		ok := false
		var prevN uint64

		for {
			total := ""
			current := ""

			err := chromedp.Run(ctx,
				badoo.wrap("wait photo", chromedp.WaitVisible(`img.js-mm-photo`)),
				badoo.wrap("extract total", chromedp.Text(`span.js-gallery-photo-total`, &total)),
				badoo.wrap("extract current", chromedp.Text(`span.js-gallery-photo-current`, &current)),
				badoo.wrap("extract photo url", chromedp.AttributeValue(`img.js-mm-photo`, "src", &url, &ok)),
			)

			if err != nil {
				return err
			}

			totalN, _ := strconv.ParseUint(total, 10, 16)
			currentN, _ := strconv.ParseUint(current, 10, 16)

			if currentN == prevN {
				badoo.wrap("go to next photo", chromedp.Click(`span.photo-gallery__link.photo-gallery__link--next.js-gallery-next`)).Do(ctx)
				continue
			}

			badoo.log.Debugf("photo %d / %d", currentN, totalN)

			if !ok {
				break
			}

			user.Photos = append(user.Photos, url)

			if currentN == totalN {
				break
			}

			prevN = currentN

			badoo.wrap("go to next photo", chromedp.Click(`span.photo-gallery__link.photo-gallery__link--next.js-gallery-next`)).Do(ctx)
		}

		return nil
	}
}

func (badoo *BadooRequester) goDating() error {
	if badoo.status == statusInited {
		return fmt.Errorf("invalid status: unauthorized")
	}

	if badoo.status != statusDating {
		badoo.status = statusDating
		return badoo.run(
			badoo.wrap("go to datings", chromedp.Navigate("https://badoo.com/encounters")),
			badoo.wrap("wait profile name", chromedp.WaitVisible(profileNamePath)),
		)
	}

	return nil
}

func (badoo *BadooRequester) goWalking() error {
	if badoo.status == statusInited {
		return fmt.Errorf("invalid status: unauthorized")
	}

	if badoo.status != statusWalking {
		badoo.status = statusWalking
		return badoo.run(
			badoo.wrap("go to search around", chromedp.Navigate("https://badoo.com/search")),
			badoo.wrap("wait first link", chromedp.WaitVisible(`a.user-card__link`)),
		)
	}

	return nil
}
