package netchange

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowsSource struct {
	id       uint64
	mu       sync.Mutex
	previous snapshot
	routes   string
	notify   func(Event)
	handles  []windows.Handle
}

var (
	windowsSourceID atomic.Uint64
	windowsSources  sync.Map
)

// Keep callback thunks process-wide: Windows callback slots cannot be freed.
var addressCallback = windows.NewCallback(func(caller unsafe.Pointer, _ uintptr, _ uint32) uintptr {
	windowsEvent(caller, false)
	return 0
})

var routeCallback = windows.NewCallback(func(caller, row unsafe.Pointer, _ uint32) uintptr {
	if row == nil || (*windows.MibIpForwardRow2)(row).DestinationPrefix.PrefixLength != 0 {
		return 0
	}

	windowsEvent(caller, true)

	return 0
})

func windowsEvent(caller unsafe.Pointer, routing bool) {
	id := *(*uint64)(caller)
	value, ok := windowsSources.Load(id)
	if !ok {
		return
	}

	s := value.(*windowsSource)

	s.mu.Lock()
	defer s.mu.Unlock()

	next, err := readSnapshot()
	if err != nil {
		return
	}

	event := s.previous.difference(next)

	if routing {
		if routes, err := readWindowsRoutes(); err == nil {
			event.Route = routes != s.routes
			s.routes = routes
		}
	}

	s.previous = next

	if event.Address || event.Link || event.Route {
		s.notify(event)
	}
}

func watch(notify func(Event)) (func(), error) {
	previous, err := readSnapshot()
	if err != nil {
		return nil, err
	}

	s := &windowsSource{id: windowsSourceID.Add(1), previous: previous, notify: notify}

	s.routes, err = readWindowsRoutes()
	if err != nil {
		return nil, err
	}

	windowsSources.Store(s.id, s)

	stop := func() {
		for _, handle := range s.handles {
			windows.CancelMibChangeNotify2(handle)
		}
		windowsSources.Delete(s.id)
	}

	var handle windows.Handle

	// IPv4/IPv6 address and interface changes are filtered by actual snapshots.
	if err := windows.NotifyUnicastIpAddressChange(
		windows.AF_UNSPEC,
		addressCallback,
		unsafe.Pointer(&s.id),
		false,
		&handle,
	); err != nil {
		stop()
		return nil, err
	}

	s.handles = append(s.handles, handle)

	if err := windows.NotifyIpInterfaceChange(
		windows.AF_UNSPEC,
		addressCallback,
		unsafe.Pointer(&s.id),
		false,
		&handle,
	); err != nil {
		stop()
		return nil, err
	}

	s.handles = append(s.handles, handle)

	if err := windows.NotifyRouteChange2(
		windows.AF_UNSPEC,
		routeCallback,
		unsafe.Pointer(&s.id),
		false,
		&handle,
	); err != nil {
		stop()
		return nil, err
	}

	s.handles = append(s.handles, handle)

	return stop, nil
}

func readWindowsRoutes() (string, error) {
	var table *windows.MibIpForwardTable2
	if err := windows.GetIpForwardTable2(windows.AF_UNSPEC, &table); err != nil {
		return "", err
	}

	defer windows.FreeMibTable(unsafe.Pointer(table))

	var routes []string
	for _, row := range table.Rows() {
		if row.DestinationPrefix.PrefixLength != 0 || row.Loopback != 0 {
			continue
		}

		address := row.NextHop.Data[:]
		if row.NextHop.Family == windows.AF_INET {
			address = address[:1]
		} else if row.NextHop.Family == windows.AF_INET6 {
			address = address[1:]
		} else {
			continue
		}

		routes = append(routes, fmt.Sprintf("%d/%d/%d/%d/%v",
			row.InterfaceLuid,
			row.InterfaceIndex,
			row.NextHop.Family,
			row.Metric,
			address,
		))
	}

	sort.Strings(routes)

	return strings.Join(routes, "\x00"), nil
}
