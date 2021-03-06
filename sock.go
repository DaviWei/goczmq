package goczmq

/*
#include "czmq.h"
#include <stdlib.h>
#include <string.h>

int Sock_connect(zsock_t *self, const char *format) {return zsock_connect(self, format, NULL);}
int Sock_disconnect(zsock_t *self, const char *format) {return zsock_disconnect(self, format, NULL);}
int Sock_bind(zsock_t *self, const char *format) {return zsock_bind(self, format, NULL);}
int Sock_unbind(zsock_t *self, const char *format) {return zsock_unbind(self, format, NULL);}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"unsafe"
)

var (
	// ErrSliceFull is returned if a []byte passed to Read was not
	// large enough to hold the contents of a message
	ErrSliceFull = errors.New("goczmq: slice full")
)

// Sock wraps the zsock_t class in CZMQ.
type Sock struct {
	zsockT       *C.struct__zsock_t
	file         string
	line         int
	zType        int
	lastClientID string
}

func init() {
	if err := os.Setenv("ZSYS_SIGHANDLER", "false"); err != nil {
		panic(err)
	}
}

// GetLastClientID returns the id of the last client you received
// a message from if the underlying socket is a ROUTER or SERVER
// socket
func (s *Sock) GetLastClientID() []byte {
	return []byte(s.lastClientID)
}

// SetLastClientID sets the client id that will be used when
// sending from a ROUTER or SERVER socket
func (s *Sock) SetLastClientID(id []byte) {
	s.lastClientID = string(id)
}

// NewSock creates a new socket.  The caller source and
// line number are passed so CZMQ can report socket leaks
// intelligently.
func NewSock(t int) *Sock {
	var s *Sock
	_, file, line, ok := runtime.Caller(1)

	if ok {
		s = &Sock{
			file:  file,
			line:  line,
			zType: t,
		}
	} else {
		s = &Sock{
			file:  "",
			line:  0,
			zType: t,
		}
	}

	s.zsockT = C.zsock_new_checked(C.int(s.zType), C.CString(s.file), C.size_t(s.line))
	return s
}

// Connect connects a socket to an endpoint
// returns an error if the connect failed.
func (s *Sock) Connect(endpoint string) error {
	rc := C.Sock_connect(s.zsockT, C.CString(endpoint))
	if rc != C.int(0) {
		return errors.New("failed")
	}
	return nil
}

// Disconnect disconnects a socket from an endpoint.  If returns
// an error if the endpoint was not found
func (s *Sock) Disconnect(endpoint string) error {
	rc := C.Sock_disconnect(s.zsockT, C.CString(endpoint))
	if int(rc) == -1 {
		return fmt.Errorf("endopint was not bound")
	}
	return nil
}

// Bind binds a socket to an endpoint.  On success returns
// the port number used for tcp transports, or 0 for other
// transports.  On failure returns a -1 for port, and an error.
func (s *Sock) Bind(endpoint string) (int, error) {
	port := C.Sock_bind(s.zsockT, C.CString(endpoint))
	if port == C.int(-1) {
		return -1, errors.New("failed")
	}
	return int(port), nil
}

// Unbind unbinds a socket from an endpoint.  If returns
// an error if the endpoint was not found
func (s *Sock) Unbind(endpoint string) error {
	rc := C.Sock_unbind(s.zsockT, C.CString(endpoint))
	if int(rc) == -1 {
		return fmt.Errorf("endopint was not bound")
	}
	return nil
}

// Attach attaches a socket to zero or more endpoints. If endpoints is not null,
// parses as list of ZeroMQ endpoints, separated by commas, and prefixed by
// '@' (to bind the socket) or '>' (to attach the socket). If the endpoint
// does not start with '@' or '>', the serverish argument determines whether
// it is used to bind (serverish = true) or connect (serverish = false)
func (s *Sock) Attach(endpoints string, serverish bool) error {
	rc := C.zsock_attach(s.zsockT, C.CString(endpoints), C._Bool(serverish))
	if rc == -1 {
		return ErrSockAttach
	}
	return nil
}

// NewPUB creates a PUB socket and calls Attach.
// The socket will Bind by default.
func NewPUB(endpoints string) (*Sock, error) {
	s := NewSock(PUB)
	return s, s.Attach(endpoints, true)
}

// NewSUB creates a SUB socket and calls Attach.
// 'subscribe' is a comma delimited list of topics to subscribe to.
// The socket will Connect by default.
func NewSUB(endpoints string, subscribe string) (*Sock, error) {
	s := NewSock(SUB)
	subscriptions := strings.Split(subscribe, ",")

	for _, topic := range subscriptions {
		s.SetSubscribe(topic)
	}

	return s, s.Attach(endpoints, false)
}

// NewREP creates a REP socket and calls Attach.
// The socket will Bind by default.
func NewREP(endpoints string) (*Sock, error) {
	s := NewSock(REP)
	return s, s.Attach(endpoints, true)
}

// NewREQ creates a REQ socket and calls Attach.
// The socket will Connect by default.
func NewREQ(endpoints string) (*Sock, error) {
	s := NewSock(REQ)
	return s, s.Attach(endpoints, false)
}

// NewPULL creates a PULL socket and calls Attach.
// The socket will Bind by default.
func NewPULL(endpoints string) (*Sock, error) {
	s := NewSock(PULL)
	return s, s.Attach(endpoints, true)
}

// NewPUSH creates a PUSH socket and calls Attach.
// The socket will Connect by default.
func NewPUSH(endpoints string) (*Sock, error) {
	s := NewSock(PUSH)
	return s, s.Attach(endpoints, false)
}

// NewROUTER creates a ROUTER socket and calls Attach.
// The socket will Bind by default.
func NewROUTER(endpoints string) (*Sock, error) {
	s := NewSock(ROUTER)
	return s, s.Attach(endpoints, true)
}

// NewDEALER creates a DEALER socket and calls Attach.
// The socket will Connect by default.
func NewDEALER(endpoints string) (*Sock, error) {
	s := NewSock(DEALER)
	return s, s.Attach(endpoints, false)
}

// NewXPUB creates an XPUB socket and calls Attach.
// The socket will Bind by default.
func NewXPUB(endpoints string) (*Sock, error) {
	s := NewSock(XPUB)
	return s, s.Attach(endpoints, true)
}

// NewXSUB creates an XSUB socket and calls Attach.
// The socket will Connect by default.
func NewXSUB(endpoints string) (*Sock, error) {
	s := NewSock(XSUB)
	return s, s.Attach(endpoints, false)
}

// NewPAIR creates a PAIR socket and calls Attach.
// The socket will Connect by default.
func NewPAIR(endpoints string) (*Sock, error) {
	s := NewSock(PAIR)
	return s, s.Attach(endpoints, false)
}

// NewSTREAM creates a STREAM socket and calls Attach.
// The socket will Connect by default.
func NewSTREAM(endpoints string) (*Sock, error) {
	s := NewSock(STREAM)
	return s, s.Attach(endpoints, false)
}

// Pollin returns true if there is a POLLIN
// event on the socket
func (s *Sock) Pollin() bool {
	return s.Events() == POLLIN
	// return C.zsock_events(unsafe.Pointer(s.zsockT)) == C.ZMQ_POLLIN
}

// Pollout returns true if there is a POLLOUT
// event on the socket
func (s *Sock) Pollout() bool {
	return s.Events() == POLLOUT
}

// SendFrame sends a byte array via the socket.  For the flags
// value, use 0 for a single message, or SNDMORE if it is
// a multi-part message
func (s *Sock) SendFrame(data []byte, flags int) error {
	var rc C.int

	if len(data) == 0 {
		frame := C.zframe_new(unsafe.Pointer(C.CString("")), C.size_t(len(data)))
		rc = C.zframe_send(&frame, unsafe.Pointer(s.zsockT), C.int(flags))
	} else {
		frame := C.zframe_new(unsafe.Pointer(&data[0]), C.size_t(len(data)))
		rc = C.zframe_send(&frame, unsafe.Pointer(s.zsockT), C.int(flags))
	}

	if rc == C.int(-1) {
		return errors.New("failed")
	}

	return nil
}

// RecvFrame reads a frame from the socket and returns it
// as a byte array, along with a more flag and and error
// (if there is an error)
func (s *Sock) RecvFrame() ([]byte, int, error) {
	frame := C.zframe_recv(unsafe.Pointer(s.zsockT))
	if frame == nil {
		return []byte{0}, 0, errors.New("failed")
	}
	dataSize := C.zframe_size(frame)
	dataPtr := C.zframe_data(frame)
	b := C.GoBytes(unsafe.Pointer(dataPtr), C.int(dataSize))
	more := C.zframe_more(frame)
	C.zframe_destroy(&frame)
	return b, int(more), nil
}

// RecvFrameNoWait receives a frame from the socket
// and returns it as a byte array if one is waiting.
// Returns an empty frame, a 0 more flag and an error
// if one is not immediately available
func (s *Sock) RecvFrameNoWait() ([]byte, int, error) {
	if !s.Pollin() {
		return []byte{0}, 0, fmt.Errorf("no frame")
	}

	return s.RecvFrame()
}

// SendMessage accepts an array of byte arrays and
// sends it as a multi-part message.
func (s *Sock) SendMessage(parts [][]byte) error {
	var f int
	numParts := len(parts)
	for i, val := range parts {
		if i == numParts-1 {
			f = 0
		} else {
			f = MORE
		}

		err := s.SendFrame(val, f)
		if err != nil {
			return err
		}
	}
	return nil
}

// RecvMessage receives a full message from the socket
// and returns it as an array of byte arrays.
func (s *Sock) RecvMessage() ([][]byte, error) {
	var msg [][]byte

	for {
		frame, flag, err := s.RecvFrame()
		if err != nil {
			return msg, err
		}
		msg = append(msg, frame)
		if flag != MORE {
			break
		}
	}
	return msg, nil
}

// Read provides an io.Reader interface to a zeromq socket
func (s *Sock) Read(p []byte) (int, error) {
	var total int
	frame, flag, err := s.RecvFrame()
	if err != nil {
		return total, err
	}

	if s.GetType() == ROUTER {
		s.lastClientID = string(frame)
	} else {
		copy(p[:], frame[:])
		total += len(frame)
	}

	for flag == MORE {
		frame, flag, err = s.RecvFrame()
		if err != nil {
			return total, err
		}
		copy(p[total:], frame[:])
		total += len(frame)
	}

	if total > len(p) {
		err = ErrSliceFull
	} else {
		err = io.EOF
	}

	return total, err
}

// Write provides an io.Writer interface to a zeromq socket
func (s *Sock) Write(p []byte) (int, error) {
	var total int
	if s.GetType() == ROUTER {
		err := s.SendFrame(s.GetLastClientID(), MORE)
		if err != nil {
			return total, err
		}
	}
	err := s.SendFrame(p, 0)
	if err != nil {
		return total, err
	}

	return len(p), nil
}

// RecvMessageNoWait receives a full message from the socket
// and returns it as an array of byte arrays if one is waiting.
// Returns an empty message and an error if one is not immediately
// available
func (s *Sock) RecvMessageNoWait() ([][]byte, error) {
	var msg [][]byte
	if !s.Pollin() {
		return msg, fmt.Errorf("no message")
	}

	for {
		frame, flag, err := s.RecvFrame()
		if err != nil {
			return msg, err
		}
		msg = append(msg, frame)
		if flag != MORE {
			break
		}
	}
	return msg, nil
}

// GetType returns the socket's type
func (s *Sock) GetType() int {
	return s.zType
}

// Destroy destroys the underlying zsockT.
func (s *Sock) Destroy() {
	C.zsock_destroy_checked(&s.zsockT, C.CString(s.file), C.size_t(s.line))
}
