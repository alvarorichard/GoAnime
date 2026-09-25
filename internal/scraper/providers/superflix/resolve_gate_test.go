package superflix

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type observedDoneContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
}

func newObservedDoneContext(ctx context.Context) *observedDoneContext {
	return &observedDoneContext{
		Context:  ctx,
		observed: make(chan struct{}),
	}
}

func (c *observedDoneContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.observed) })
	return c.Context.Done()
}

// cancelOnSecondErrContext makes the cancellation land after lock has passed
// its initial context check and acquired the available token.
// It depends on exactly two Err() calls; if that sequence changes, the test
// must fail instead of silently missing the acquisition race.
type cancelOnSecondErrContext struct {
	context.Context
	done  chan struct{}
	calls atomic.Int32
	once  sync.Once
}

func newCancelOnSecondErrContext() *cancelOnSecondErrContext {
	return &cancelOnSecondErrContext{
		Context: context.Background(),
		done:    make(chan struct{}),
	}
}

func (c *cancelOnSecondErrContext) Done() <-chan struct{} { return c.done }

func (c *cancelOnSecondErrContext) Err() error {
	if c.calls.Add(1) == 1 {
		return nil
	}
	c.once.Do(func() { close(c.done) })
	return context.Canceled
}

func TestContextGateCanceledWaiterDoesNotConsumeToken(t *testing.T) {
	var gate contextGate
	if err := gate.lock(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	waitingCtx := newObservedDoneContext(ctx)
	result := make(chan error, 1)
	go func() { result <- gate.lock(waitingCtx) }()

	select {
	case <-waitingCtx.observed:
	case <-time.After(time.Second):
		cancel()
		gate.unlock()
		t.Fatal("waiter did not reach the cancellable wait")
	}
	cancel()

	select {
	case err := <-result:
		if err != context.Canceled {
			t.Fatalf("lock error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		gate.unlock()
		t.Fatal("canceled waiter did not return promptly")
	}

	gate.unlock()
	if err := gate.lock(context.Background()); err != nil {
		t.Fatalf("gate was not reusable after cancellation: %v", err)
	}
	gate.unlock()
}

func TestContextGateReturnsTokenWhenCancellationRacesAcquisition(t *testing.T) {
	var gate contextGate
	ctx := newCancelOnSecondErrContext()
	if err := gate.lock(ctx); err != context.Canceled {
		t.Fatalf("lock error = %v, want context.Canceled", err)
	}
	// This context deliberately relies on lock() checking Err() once before
	// waiting and once after taking the token. Keep the assertion so a refactor
	// cannot silently move the cancellation before the acquisition branch.
	if calls := ctx.calls.Load(); calls != 2 {
		t.Fatalf("lock checked context Err() %d times, want exactly 2 for this race", calls)
	}

	if err := gate.lock(context.Background()); err != nil {
		t.Fatalf("token was lost in the cancellation race: %v", err)
	}
	gate.unlock()
}

func TestContextGateUnlockWithoutLockPanics(t *testing.T) {
	var gate contextGate
	defer func() {
		if recover() == nil {
			t.Fatal("unlocking an unlocked gate must panic immediately")
		}
	}()
	gate.unlock()
}

func TestKeyedGateGroupSerializesSameKeyAndAllowsDifferentKeys(t *testing.T) {
	var group keyedGateGroup
	firstRelease, err := group.lock(context.Background(), "episode:1")
	if err != nil {
		t.Fatal(err)
	}

	waitingCtx := newObservedDoneContext(context.Background())
	secondResult := make(chan error, 1)
	go func() {
		release, lockErr := group.lock(waitingCtx, "episode:1")
		if release != nil {
			release()
		}
		secondResult <- lockErr
	}()
	select {
	case <-waitingCtx.observed:
	case <-time.After(time.Second):
		firstRelease()
		t.Fatal("same-key waiter did not reach the gate")
	}

	otherRelease, err := group.lock(context.Background(), "episode:2")
	if err != nil {
		firstRelease()
		t.Fatalf("a different key was serialized: %v", err)
	}
	otherRelease()

	select {
	case err := <-secondResult:
		firstRelease()
		t.Fatalf("same-key waiter passed before release: %v", err)
	default:
	}

	firstRelease()
	select {
	case err := <-secondResult:
		if err != nil {
			t.Fatalf("same-key waiter failed after release: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("same-key waiter did not proceed after release")
	}

	group.mu.Lock()
	remaining := len(group.gates)
	group.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("keyed gate map retained %d entries after all users exited", remaining)
	}
}

func TestKeyedGateGroupCanceledWaiterCleansUp(t *testing.T) {
	var group keyedGateGroup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := group.lock(ctx, "episode:cancelled"); err != context.Canceled {
		t.Fatalf("lock error = %v, want context.Canceled", err)
	}

	group.mu.Lock()
	remaining := len(group.gates)
	group.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("canceled caller left %d keyed gate entries", remaining)
	}
}
