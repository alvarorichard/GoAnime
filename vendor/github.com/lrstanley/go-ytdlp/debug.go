// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"context"
	"log/slog"
	"os"
	"strconv"
)

var debugLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

func debug(ctx context.Context, msg string, args ...any) {
	debug, _ := strconv.ParseBool(os.Getenv("YTDLP_DEBUG"))
	if !debug {
		return
	}
	debugLogger.DebugContext(ctx, msg, args...) //nolint:sloglint
}
