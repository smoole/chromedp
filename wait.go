package chromedp

import (
	"context"
	"sync"
)

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
