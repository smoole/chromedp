package chromedp

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
)

func CookieParamsFromCookies(cookies []*network.Cookie) []*network.CookieParam {
	ret := []*network.CookieParam{}
	for _, c := range cookies {
		expr := cdp.TimeSinceEpoch(time.Unix(int64(c.Expires), 0))

		ret = append(ret, &network.CookieParam{
			Name:     c.Name,
			Value:    c.Value,
			URL:      "",
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
			Expires:  &expr,
		})
	}
	return ret
}

func SetCookies(cookies []*network.Cookie) Action {
	return ActionFunc(func(ctx context.Context) error {
		params := CookieParamsFromCookies(cookies)
		if err := network.SetCookies(params).Do(ctx); err != nil {
			return err
		}
		return nil
	})
}

func GetCookies(cookies *[]*network.Cookie) Action {
	if cookies == nil {
		panic("cookies cannot be nil")
	}

	return ActionFunc(func(ctx context.Context) error {
		ac, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return err
		}
		*cookies = ac
		return nil
	})
}
