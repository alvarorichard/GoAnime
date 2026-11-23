//go:build gomock || generate

package quic

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_send_conn_test.go github.com/enetx/quic-go SendConn"
type SendConn = sendConn

//go:generate sh -c "go tool mockgen -typed -package quic -self_package github.com/enetx/quic-go -mock_names Conn=MockRawConn -destination mock_raw_conn_test.go github.com/enetx/quic-go/internal/udp Conn"

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_sender_test.go github.com/enetx/quic-go Sender"
type Sender = sender

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_stream_sender_test.go github.com/enetx/quic-go StreamSender"
type StreamSender = streamSender

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_stream_control_frame_getter_test.go github.com/enetx/quic-go StreamControlFrameGetter"
type StreamControlFrameGetter = streamControlFrameGetter

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_stream_frame_getter_test.go github.com/enetx/quic-go StreamFrameGetter"
type StreamFrameGetter = streamFrameGetter

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_frame_source_test.go github.com/enetx/quic-go FrameSource"
type FrameSource = frameSource

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_ack_frame_source_test.go github.com/enetx/quic-go AckFrameSource"
type AckFrameSource = ackFrameSource

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_sealing_manager_test.go github.com/enetx/quic-go SealingManager"
type SealingManager = sealingManager

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_unpacker_test.go github.com/enetx/quic-go Unpacker"
type Unpacker = unpacker

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_packer_test.go github.com/enetx/quic-go Packer"
type Packer = packer

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_conn_runner_test.go github.com/enetx/quic-go ConnRunner"
type ConnRunner = connRunner

//go:generate sh -c "go tool mockgen -typed -build_flags=\"-tags=gomock\" -package quic -self_package github.com/enetx/quic-go -destination mock_packet_handler_test.go github.com/enetx/quic-go PacketHandler"
type PacketHandler = packetHandler
