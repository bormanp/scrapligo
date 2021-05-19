package transport

import (
	"errors"
	"time"
)

// constants for basic transport values.
const (
	ReadSize              = 65_535
	SystemTransportName   = "system"
	StandardTransportName = "standard"
)

// ErrTransportFailure error for EOF/failure reading from the transport.
var ErrTransportFailure = errors.New("error reading from transport, cannot continue")

// ErrUnknownTransport error for when user provides an unknown/unsupported transport name.
var ErrUnknownTransport = errors.New("unknown transport provided")

// ErrTransportTimeout error for transport operations timing out.
var ErrTransportTimeout = errors.New("transport operation timed out")

// ErrKeyVerificationFailed ssh key verification failure.
var ErrKeyVerificationFailed = errors.New("ssh key verification failed")

// BaseTransportArgs struct for attributes that are required for any transport type.
type BaseTransportArgs struct {
	Host         string
	Port         int
	AuthUsername string
	// passed as pointers so they can be modified at the driver layer
	TimeoutSocket    *time.Duration
	TimeoutTransport *time.Duration
	// LoggingUID       string
	// TransportOptions; not sure how to handle this yet as it is very... vague and... this is go... :)
}

type transportResult struct {
	result []byte
	error  error
}

// BaseTransport interface defining required methods for any transport type.
type BaseTransport interface {
	Open() error
	OpenNetconf() error
	Close() error
	IsAlive() bool
	Read() ([]byte, error)
	Write([]byte) error
	FormatLogMessage(string, string) string
}

func transportTimeout(
	timeout time.Duration,
	f func() *transportResult,
) ([]byte, error) {
	c := make(chan *transportResult, 1)

	go func() {
		r := f()
		c <- r
		close(c)
	}()

	timer := time.NewTimer(timeout)

	select {
	case r := <-c:
		return r.result, r.error
	case <-timer.C:
		return make([]byte, 0), ErrTransportTimeout
	}
}