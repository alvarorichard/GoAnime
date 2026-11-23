package http3

import (
	"errors"
	"io"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
)

// startInitialSettings has one owner, including when the transport creates the
// client from OnApplicationReady. A blocked write is handed off with its original
// stream and bytes; a second critical control stream is never opened.
func (c *ClientConn) startInitialSettings(datagrams bool, settings g.MapOrd[uint64, uint64], nonblocking bool) {
	finish := func(err error) {
		if err == nil && c.rawConn.clientQPACK != nil {
			err = c.rawConn.clientQPACK.stageEagerTypes()
		}
		c.initialSettingsErr = err
		close(c.initialSettingsWritten)
		if err != nil && !errors.Is(err, quic.Err0RTTRejected) {
			if c.logger != nil {
				c.logger.Debug("setting up connection failed", "error", err)
			}
			// A synchronous CloseWithError would wait on the send loop invoking us.
			go c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeInternalError), "setting up HTTP/3 control stream failed")
		}
	}
	if err := c.wireConfig.Validate(); err != nil {
		finish(&RequestPreparationError{err})
		return
	}
	var err error
	settings, err = connectionSettingsGrease(settings, c.wireConfig.SettingsGrease)
	if err != nil {
		finish(&RequestPreparationError{err})
		return
	}
	if c.wireConfig.ControlGreaseFrames != nil {
		frames, err := c.wireConfig.ControlGreaseFrames()
		if err == nil {
			c.controlWriter.initial, err = appendGreaseFrames(frames)
		}
		if err != nil {
			finish(&RequestPreparationError{err})
			return
		}
	}
	stream, data, err := c.rawConn.prepareControlStream(initialClientSettings(datagrams, settings))
	if err == nil && c.rawConn.clientQPACK != nil {
		err = c.rawConn.clientQPACK.prepare()
	}
	if err != nil {
		finish(err)
		return
	}
	if nonblocking {
		// Keep reserved control frames next to SETTINGS in the initial stream write.
		recordWireFrames(c.qlogger, stream.StreamID(), c.controlWriter.initial)
		data = append(data, c.controlWriter.initial...)
		c.controlWriter.initial = nil
		err = stream.TryWriteAll(data)
		if !errors.Is(err, quic.ErrWouldBlock) {
			if err == nil {
				c.controlWriter.attachStream(stream)
				go c.controlWriter.run(stream)
			}
			finish(err)
			return
		}
	}
	write := func() {
		n, err := stream.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		if err == nil {
			c.controlWriter.attachStream(stream)
			go c.controlWriter.run(stream)
		}
		finish(err)
	}
	if nonblocking {
		go write()
	} else {
		write()
	}
}
