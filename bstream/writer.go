package bstream

import "sync/atomic"

type Writer struct {
	*Stream
	overwriteOldData bool
}

func (s *Writer) Close() error {
	s.wcancel()
	return nil
}

func (s *Writer) Write(b []byte) (int, error) {
BEGIN:
	select {
	case <-s.wctx.Done():
		return 0, ErrWriterClosed
	default:
	}

	dataSize := uint64(len(b))
	h := *s.head
	t := *s.tail
	var availSize uint64
	if h < t {
		availSize = s.bufSize - (t - h) - 1
	} else if h > t {
		availSize = h - t - 1
	} else {
		availSize = s.bufSize - 1
	}

	if availSize < dataSize {
		if !s.overwriteOldData {
			return 0, ErrNotEnoughMemory
		}

		needOffset := dataSize - availSize
		newHead := (h + needOffset) % s.bufSize
		if !atomic.CompareAndSwapUint64(s.head, h, newHead) {
			goto BEGIN
		}
	}

	part1 := min(dataSize, s.bufSize-t)
	part2 := dataSize - part1
	copy(s.b[t:t+part1], b[:part1])
	copy(s.b[:part2], b[part1:part1+part2])

	(*s.tail) = (t + dataSize) % s.bufSize
	return int(dataSize), nil
}
