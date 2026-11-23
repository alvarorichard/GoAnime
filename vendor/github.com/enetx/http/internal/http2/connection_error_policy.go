package http2

// closeForPeerError applies application policy before waking any requests.
// Browser-specific codes, origin state and retry rules belong to the caller.
func (rl *clientConnReadLoop) closeForPeerError(frame FrameType, id uint32, code ErrCode) bool {
	cc := rl.cc
	fp := cc.fingerprint
	if fp == nil || fp.ConnectionErrorPolicy == nil || !fp.ConnectionErrorPolicy(ConnectionErrorInfo{
		Conn: cc.tconn, FrameType: frame, StreamID: id, Code: code,
	}) {
		return false
	}
	cc.SetDoNotReuse()
	cc.t.connPool.MarkDead(cc)
	return true
}
