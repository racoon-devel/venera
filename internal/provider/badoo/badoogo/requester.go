package badoogo

import (
	"context"
	"regexp"
	"time"

	"github.com/ccding/go-logging/logging"
	"github.com/chromedp/chromedp"
)

const (
	opTimeout    = 30 * time.Second
	statusInited = iota
	statusAuthorized
	statusDating
	statusWalking
)

const (
	loginButtonPath = `#page > div.page__simple-wrap > div.page__content > section > div > div > div.sign-page__form.js-core-events-container > form > div.new-form__actions > div > div:nth-child(1) > button`
	sidebarUserPath = `a.sidebar-info__name`

	profileNamePath   = `span.profile-header__name`
	profileAgePath    = `span.profile-header__age`
	profileJobPath    = `#mm_cc > div.encounters-card__inner > div > div > div.scroll__inner > div > div.encounters-card__header.js-profile-header-container.js-core-events-container > div > div.profile-header__metrics > div > div > div`
	profileAboutPath  = `#mm_cc > div.encounters-card__inner > div > div > div.scroll__inner > div > div.encounters-card__info > div.encounters-card__section.js-profile-about-me-container.js-core-events-container > div > div.profile-section__content > div > p`
	profileTargetPath = `#app_c > div > div.profile__content.js-profile-layout-animated-content > div.profile__info > section.profile__main-info > div.profile__section.js-profile-iht-container > div > div.profile-section__content`
	cardLinkPath      = `#mm_cc > div.encounters-card__inner > div > div > div.scroll__inner > div > div.encounters-card__header.js-profile-header-container.js-core-events-container > div > div.profile-header__inner > div.profile-header__info > h1 > span.b-link.js-profile-header-name.js-hp-view-element`
	placeInfoPath     = `#app_c > div > div.profile__content.js-profile-layout-animated-content > div.profile__info > section.profile__main-info > div.profile__section.js-profile-location-container.js-core-events-container > div > div.profile-section__title > h2`

	likeButtonPath = `#app_c > div > div.profile__header.js-profile-header-container.js-core-events-container > header > div.profile-header__vote.js-profile-header-buttons > div.profile-header__vote-item.profile-header__vote-item--yes.js-tutorial-first-yes > div`
	passButtonPath = `#app_c > div > div.profile__header.js-profile-header-container.js-core-events-container > header > div.profile-header__vote.js-profile-header-buttons > div.profile-header__vote-item.profile-header__vote-item--no.js-tutorial-first-no > div`
)

type BadooRequester struct {
	ctx    context.Context
	cancel context.CancelFunc
	status int

	log       *logging.Logger
	ageExpr   *regexp.Regexp
	urlExpr   *regexp.Regexp
	debugMode bool
}

func NewBadooRequester(ctx context.Context, log *logging.Logger) *BadooRequester {
	badoo := newRequester(log)
	badoo.ctx, badoo.cancel = chromedp.NewContext(ctx)
	return badoo
}

func (badoo *BadooRequester) Close() {
	chromedp.Cancel(badoo.ctx)
}

func newRequester(log *logging.Logger) *BadooRequester {
	return &BadooRequester{
		log:     log,
		ageExpr: regexp.MustCompile(`[\d]+`),
		urlExpr: regexp.MustCompile(`url\(([^)]+)`),
	}
}
