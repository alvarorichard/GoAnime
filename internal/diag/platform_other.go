//go:build !linux && !darwin && !windows

package diag

import "context"

func platformInfo(context.Context, Options) []Row { return nil }
