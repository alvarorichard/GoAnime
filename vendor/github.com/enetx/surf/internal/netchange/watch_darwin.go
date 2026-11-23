package netchange

import (
	"os"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func watch(notify func(Event)) (func(), error) {
	fd, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, unix.AF_UNSPEC)
	if err != nil {
		return nil, err
	}

	unix.CloseOnExec(fd)
	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return nil, err
	}

	file := os.NewFile(uintptr(fd), "surf-network-notifications")
	previous, err := readSnapshot()
	if err != nil {
		file.Close()
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		buffer := make([]byte, 65536)
		for {
			n, err := file.Read(buffer)
			if err != nil {
				return
			}

			messages, err := route.ParseRIB(unix.RTM_VERSION, buffer[:n])
			if err != nil {
				continue
			}

			event := Event{}

			for _, message := range messages {
				if m, ok := message.(*route.RouteMessage); ok &&
					(m.Type == unix.RTM_ADD || m.Type == unix.RTM_DELETE || m.Type == unix.RTM_CHANGE) {
					switch dst := m.Addrs[unix.RTAX_DST].(type) {
					case *route.Inet4Addr:
						event.Route = event.Route || dst.IP == [4]byte{}
					case *route.Inet6Addr:
						event.Route = event.Route || dst.IP == [16]byte{}
					}
				}
			}

			next, err := readSnapshot()
			if err != nil {
				continue
			}

			difference := previous.difference(next)

			event.Address, event.Link = difference.Address, difference.Link
			event.IPv6Only = difference.IPv6Only
			previous = next

			if event.Address || event.Link || event.Route {
				notify(event)
			}
		}
	}()

	return func() { file.Close(); <-done }, nil
}
