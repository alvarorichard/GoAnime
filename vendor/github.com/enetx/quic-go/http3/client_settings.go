package http3

import "context"

// Queue the control stream before making a request stream writable. Starting
// both from independent goroutines lets HEADERS overtake our initial SETTINGS.
// This orders local writes; it does not assume delivery order between streams.
func (c *ClientConn) waitInitialSettings(ctx context.Context) error {
	if c.initialSettingsWritten == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.conn.Context().Done():
		select {
		case <-c.initialSettingsWritten:
			if c.initialSettingsErr != nil {
				return c.initialSettingsErr
			}
		default:
		}
		return context.Cause(c.conn.Context())
	case <-c.initialSettingsWritten:
		return c.initialSettingsErr
	}
}
