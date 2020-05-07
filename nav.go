package chromedp

import (
	"context"
	"errors"
	"time"

	"github.com/chromedp/cdproto/page"
)

type navOptions struct {
	wait func(ctx context.Context, ev interface{}) error
}

type NavigateOption = func(*navOptions)

func NavigateWaitFunc(wait func(ctx context.Context, ev interface{}) error) NavigateOption {
	return func(opts *navOptions) {
		opts.wait = wait
	}
}

func NavigateNoWait(opts *navOptions) {
	opts.wait = nil
}

func NavigateWaitEventLoadEventFired(opts *navOptions) {
	opts.wait = func(ctx context.Context, ev interface{}) error {
		if _, ok := ev.(*page.EventLoadEventFired); ok {
			return nil
		}
		return ErrNotMatch
	}
}

func NavigateWaitEventFrameNavigated(opts *navOptions) {
	opts.wait = func(ctx context.Context, ev interface{}) error {
		if _, ok := ev.(*page.EventFrameNavigated); ok {
			return nil
		}
		return ErrNotMatch
	}
}

func WaitNavigate(opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		return waitNavEvent(ctx, opts...)
	})
}

// NavigateAction are actions that manipulate the navigation of the browser.
type NavigateAction Action

// Navigate is an action that navigates the current frame.
func Navigate(urlstr string, opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		_, _, _, err := page.Navigate(urlstr).Do(ctx)
		if err != nil {
			return err
		}
		return waitNavEvent(ctx, opts...)
	})
}

// waitLoaded blocks until a target receives a Page.loadEventFired.
func waitNavEvent(ctx context.Context, opts ...NavigateOption) error {
	options := &navOptions{}
	// set default wait func
	NavigateWaitEventLoadEventFired(options)
	// override
	for _, o := range opts {
		o(options)
	}

	// If no wait, just return
	if options.wait == nil {
		return nil
	}

	// TODO: this function is inherently racy, as we don't run ListenTarget
	// until after the navigate action is fired. For example, adding
	// time.Sleep(time.Second) at the top of this body makes most tests hang
	// forever, as they miss the load event.
	//
	// However, setting up the listener before firing the navigate action is
	// also racy, as we might get a load event from a previous navigate.
	//
	// For now, the second race seems much more common in real scenarios, so
	// keep the first approach. Is there a better way to deal with this?
	ch := make(chan struct{})
	lctx, cancel := context.WithCancel(ctx)
	ListenTarget(lctx, func(ev interface{}) {
		if err := options.wait(ctx, ev); err == nil {
			cancel()
			close(ch)
		}
	})
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NavigationEntries is an action that retrieves the page's navigation history
// entries.
func NavigationEntries(currentIndex *int64, entries *[]*page.NavigationEntry) NavigateAction {
	if currentIndex == nil || entries == nil {
		panic("currentIndex and entries cannot be nil")
	}

	return ActionFunc(func(ctx context.Context) error {
		var err error
		*currentIndex, *entries, err = page.GetNavigationHistory().Do(ctx)
		return err
	})
}

// NavigateToHistoryEntry is an action to navigate to the specified navigation
// entry.
func NavigateToHistoryEntry(entryID int64, opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		if err := page.NavigateToHistoryEntry(entryID).Do(ctx); err != nil {
			return err
		}
		return waitNavEvent(ctx, opts...)
	})
}

// NavigateBack is an action that navigates the current frame backwards in its
// history.
func NavigateBack(opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		cur, entries, err := page.GetNavigationHistory().Do(ctx)
		if err != nil {
			return err
		}

		if cur <= 0 || cur > int64(len(entries)-1) {
			return errors.New("invalid navigation entry")
		}

		entryID := entries[cur-1].ID
		if err := page.NavigateToHistoryEntry(entryID).Do(ctx); err != nil {
			return err
		}
		return waitNavEvent(ctx, opts...)
	})
}

// NavigateForward is an action that navigates the current frame forwards in
// its history.
func NavigateForward(opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		cur, entries, err := page.GetNavigationHistory().Do(ctx)
		if err != nil {
			return err
		}

		if cur < 0 || cur >= int64(len(entries)-1) {
			return errors.New("invalid navigation entry")
		}

		entryID := entries[cur+1].ID
		if err := page.NavigateToHistoryEntry(entryID).Do(ctx); err != nil {
			return err
		}
		return waitNavEvent(ctx, opts...)
	})
}

// Reload is an action that reloads the current page.
func Reload(opts ...NavigateOption) NavigateAction {
	return ActionFunc(func(ctx context.Context) error {
		if err := page.Reload().Do(ctx); err != nil {
			return err
		}
		return waitNavEvent(ctx, opts...)
	})
}

// Stop is an action that stops all navigation and pending resource retrieval.
func Stop() NavigateAction {
	return page.StopLoading()
}

// CaptureScreenshot is an action that captures/takes a screenshot of the
// current browser viewport.
//
// See the Screenshot action to take a screenshot of a specific element.
//
// See the 'screenshot' example in the https://github.com/chromedp/examples
// project for an example of taking a screenshot of the entire page.
func CaptureScreenshot(res *[]byte) Action {
	if res == nil {
		panic("res cannot be nil")
	}

	return ActionFunc(func(ctx context.Context) error {
		var err error
		*res, err = page.CaptureScreenshot().Do(ctx)
		return err
	})
}

// Location is an action that retrieves the document location.
func Location(urlstr *string) Action {
	if urlstr == nil {
		panic("urlstr cannot be nil")
	}
	return EvaluateAsDevTools(`document.location.toString()`, urlstr)
}

// Title is an action that retrieves the document title.
func Title(title *string) Action {
	if title == nil {
		panic("title cannot be nil")
	}
	return EvaluateAsDevTools(`document.title`, title)
}

func WaitNotLocation(not string, ret *string) Action {
	return ActionFunc(func(ctx context.Context) error {
		urlstr := ""
		for {
			tm := time.NewTimer(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				tm.Stop()
				return ctx.Err()
			case <-tm.C:
			}

			Location(&urlstr).Do(ctx)
			if urlstr != not {
				if ret != nil {
					*ret = urlstr
				}
				return nil
			}
		}
	})
}

func WaitLocationChanged(ret *string) Action {
	return ActionFunc(func(ctx context.Context) error {
		init := ""
		Location(&init).Do(ctx)

		return WaitNotLocation(init, ret).Do(ctx)
	})
}
