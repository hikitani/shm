package bstream

import "sync/atomic"

type Reader struct {
	*Stream
}

func (s *Reader) Close() error {
	s.rcancel()
	return nil
}

func (s *Reader) Read(b []byte) (int, error) {
BEGIN:
	select {
	case <-s.rctx.Done():
		return 0, ErrReaderClosed
	default:
	}

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
	used := (s.bufSize - 1) - availSize
	if used == 0 {
		return 0, ErrStreamEmpty
	}
	dataSize := min(uint64(len(b)), used)

	part1 := min(dataSize, s.bufSize-h)
	part2 := dataSize - part1
	copy(b[:part1], s.b[h:h+part1])
	copy(b[part1:part1+part2], b[:part2])

	newHead := (h + dataSize) % s.bufSize
	if !atomic.CompareAndSwapUint64(s.head, h, newHead) {
		goto BEGIN
	}
	return int(dataSize), nil
}
