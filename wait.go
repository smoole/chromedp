package chromedp

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrContinueWait = errors.New("ErrContinueWait")
)

func WaitUntil(f func(ctx context.Context) error) Action {
	return ActionFunc(func(ctx context.Context) error {
		for {
			tm := time.NewTimer(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				tm.Stop()
				return ctx.Err()
			case <-tm.C:
			}

			err := f(ctx)
			if err == ErrContinueWait {
				continue
			}
			return err
		}
	})
}

func IntervalRun(interval time.Duration, action Action) Action {
	return ActionFunc(func(ctx context.Context) error {
		tm := time.NewTimer(interval)
		for {
			select {
			case <-ctx.Done():
				tm.Stop()
				return ctx.Err()
			case <-tm.C:
				if err := action.Do(ctx); err != nil {
					return err
				}
				tm = time.NewTimer(interval)
			}
		}
	})
}

func WaitOneOf(waitIdx *int, actions ...Action) Action {
	if len(actions) == 0 {
		panic("actions cannot be empty")
	}

	return ActionFunc(func(ctx context.Context) error {
		wg := &sync.WaitGroup{}
		defer wg.Wait()

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		type ret struct {
			idx int
			err error
		}
		retC := make(chan ret)

		for idx := 0; idx < len(actions); idx++ {
			action := actions[idx]

			wg.Add(1)
			go func(idx int) {
				defer wg.Done()

				err := action.Do(ctx)
				select {
				case retC <- ret{
					idx: idx,
					err: err,
				}:
				case <-ctx.Done():
				}
			}(idx)
		}

		select {
		case r := <-retC:
			if r.err != nil {
				return r.err
			}

			if waitIdx != nil {
				*waitIdx = r.idx
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}
