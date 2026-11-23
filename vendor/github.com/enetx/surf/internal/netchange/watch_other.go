//go:build !linux && !darwin && !windows

package netchange

import "errors"

func watch(func(Event)) (func(), error) { return nil, errors.ErrUnsupported }
