// Package netchange delivers native network notifications to built clients.
package netchange

import (
	"net/netip"
	"slices"
	"sync"
	"time"
)

// Event distinguishes address changes from link and routing notifications.
// TemporaryIPv6Only is true only when every changed address is known temporary.
type Event struct {
	Address, Link, Route bool
	TemporaryIPv6Only    bool
	IPv6Only             bool
	Identity             bool
	// IdentityRoute requests local identity refresh for a global unicast route.
	// Neighbors list Ethernet neighbors that acquired a previously unknown MAC.
	// They are hints for identity readers, not delivered pool-change flags.
	IdentityRoute bool
	Neighbors     []netip.Addr
}

func (e Event) merge(other Event) Event {
	return Event{
		Address:       e.Address || other.Address,
		Link:          e.Link || other.Link,
		Route:         e.Route || other.Route,
		Identity:      e.Identity || other.Identity,
		IdentityRoute: e.IdentityRoute || other.IdentityRoute,
		Neighbors:     append(slices.Clone(e.Neighbors), other.Neighbors...),
		TemporaryIPv6Only: (e.Address || other.Address) && (!e.Address || e.TemporaryIPv6Only) &&
			(!other.Address || other.TemporaryIPv6Only),
		IPv6Only: (e.Address || other.Address) && (!e.Address || e.IPv6Only) && (!other.Address || other.IPv6Only),
	}
}

// Subscription owns a serial callback and its coalescing window.
// Close waits for callback completion. A callback must not call Close itself.
type Subscription struct {
	mu       sync.Mutex
	pending  Event
	wake     chan struct{}
	stop     chan struct{}
	done     chan struct{}
	once     sync.Once
	delay    time.Duration
	notify   func(Event)
	identity identityState
}

var monitor struct {
	sync.Mutex
	epoch       uint64
	stop        func()
	subscribers map[*Subscription]struct{}
}

// Subscribe shares one native source. Delay coalesces from the first event;
// later notifications merge into that window without extending it.
func Subscribe(delay time.Duration, notify func(Event), identity ...func([]netip.Addr) (string, error)) (*Subscription, error) {
	s := &Subscription{
		wake: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{}),
		delay: delay, notify: notify,
	}

	monitor.Lock()
	defer monitor.Unlock()

	if monitor.stop == nil {
		monitor.epoch++
		epoch := monitor.epoch
		stop, err := watch(func(event Event) { deliver(epoch, event) })
		if err != nil {
			return nil, err
		}

		monitor.stop = stop
	}

	if monitor.subscribers == nil {
		monitor.subscribers = make(map[*Subscription]struct{})
	}

	if len(identity) != 0 {
		s.identity.read = identity[0]
		s.identity.refresh(nil)
	}

	monitor.subscribers[s] = struct{}{}

	go s.run()

	return s, nil
}

func deliver(epoch uint64, event Event) {
	monitor.Lock()
	defer monitor.Unlock()

	if monitor.epoch != epoch {
		return
	}

	for s := range monitor.subscribers {
		s.mu.Lock()
		s.pending = s.pending.merge(event)
		s.mu.Unlock()

		select {
		case s.wake <- struct{}{}:
		default:
		}
	}
}

func (s *Subscription) run() {
	defer close(s.done)

	for {
		select {
		case <-s.stop:
			return
		case <-s.wake:
		}

		if s.delay > 0 {
			timer := time.NewTimer(s.delay)
			select {
			case <-s.stop:
				timer.Stop()
				return
			case <-timer.C:
			}
		}

		s.mu.Lock()
		event := s.pending
		s.pending = Event{}

		select {
		case <-s.wake:
		default:
		}

		s.mu.Unlock()

		event = s.identity.observe(event)

		if event.Address || event.Link || event.Route || event.Identity {
			s.notify(event)
		}
	}
}

// Close unsubscribes and releases the native source after its final subscriber.
// It is safe to call more than once.
func (s *Subscription) Close() {
	if s == nil {
		return
	}

	s.once.Do(func() {
		monitor.Lock()

		delete(monitor.subscribers, s)

		var stop func()
		if len(monitor.subscribers) == 0 {
			stop, monitor.stop = monitor.stop, nil
			monitor.epoch++
		}

		monitor.Unlock()

		close(s.stop)
		if stop != nil {
			stop()
		}
		<-s.done
	})
}
