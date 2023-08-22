package bstream

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

const BeatInterval = 5 * time.Second

var (
	ErrNotEnoughMemory  = errors.New("not enough memory")
	ErrStreamEmpty      = errors.New("stream is empty")
	ErrWriterClosed     = errors.New("stream writer closed")
	ErrReaderClosed     = errors.New("stream reader closed")
	ErrStreamWriterBusy = errors.New("stream writer is busy")
	ErrStreamReaderBusy = errors.New("stream reader is busy")
)

type Stream struct {
	wctx    context.Context
	wcancel context.CancelFunc
	rctx    context.Context
	rcancel context.CancelFunc

	b       []byte
	bufSize uint64

	producerLock *uint32
	producerBeat *uint64
	consumerLock *uint32
	consumerBeat *uint64
	head         *uint64
	tail         *uint64
	headCycle    *uint64
	tailCycle    *uint64
}

func (s *Stream) Writer() (*Writer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if !tryBufferLock(ctx, s.producerLock, s.producerBeat) {
		cancel()
		return nil, ErrStreamWriterBusy
	}

	s.wctx = ctx
	s.wcancel = cancel
	return &Writer{Stream: s}, nil
}

func (s *Stream) Reader() (*Reader, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if !tryBufferLock(ctx, s.consumerLock, s.consumerBeat) {
		cancel()
		return nil, ErrStreamReaderBusy
	}

	s.rctx = ctx
	s.rcancel = cancel
	return &Reader{Stream: s}, nil
}

func New(buffer []byte) (*Stream, error) {
	s := &Stream{}
	const minSize = unsafe.Sizeof(*s.producerLock) +
		unsafe.Sizeof(*s.producerBeat) +
		unsafe.Sizeof(*s.consumerLock) +
		unsafe.Sizeof(*s.consumerBeat) +
		unsafe.Sizeof(*s.head) +
		unsafe.Sizeof(*s.tail) +
		unsafe.Sizeof(*s.headCycle) +
		unsafe.Sizeof(*s.tailCycle)
	if len(buffer) < int(minSize) {
		return nil, fmt.Errorf("need %d bytes for reservation", minSize)
	}

	ptr := unsafe.Pointer(unsafe.SliceData(buffer))

	var offset uintptr
	s.producerLock = (*uint32)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.producerLock)
	s.producerBeat = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.producerBeat)
	s.consumerLock = (*uint32)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.consumerLock)
	s.consumerBeat = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.consumerBeat)
	s.head = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.head)
	s.tail = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.tail)
	s.headCycle = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.headCycle)
	s.tailCycle = (*uint64)(unsafe.Add(ptr, offset))
	offset += unsafe.Sizeof(*s.tailCycle)

	s.b = buffer[offset:]
	s.bufSize = uint64(len(s.b))
	return s, nil
}

func tryBufferLock(ctx context.Context, lock *uint32, beat *uint64) (ok bool) {
	defer func() {
		if ok {
			go heartBeat(ctx, beat)
		}
	}()

	if atomic.CompareAndSwapUint32(lock, 0, 1) {
		return true
	}

	oldBeat := atomic.LoadUint64(beat)
	time.Sleep(2 * BeatInterval)

	if atomic.CompareAndSwapUint64(beat, oldBeat, 0) {
		atomic.StoreUint32(lock, 1)
		return true
	}

	return false
}

func heartBeat(ctx context.Context, beat *uint64) {
	atomic.StoreUint64(beat, 0)

	t := time.NewTicker(BeatInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			atomic.AddUint64(beat, 1)
		case <-ctx.Done():
			atomic.StoreUint64(beat, 0)
			return
		}
	}
}
