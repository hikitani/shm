package bstream

type Writer struct {
	*Stream
}

func (s *Writer) Close() error {
	s.wcancel()
	return nil
}

func (s *Writer) Write(b []byte) (int, error) {
	select {
	case <-s.wctx.Done():
		return 0, ErrWriterClosed
	default:
	}

	dataSize := uint64(len(b))
	h := *s.head
	hc := *s.headCycle
	t := *s.tail
	tc := *s.tailCycle
	var availSize uint64
	if h < t {
		availSize = s.bufSize - (t - h)
	} else if h > t {
		availSize = h - t
	} else {
		if hc == tc {
			availSize = s.bufSize
		} else if tc > hc {
			availSize = 0
		}
	}

	if availSize < dataSize {
		return 0, ErrNotEnoughMemory
	}

	part1 := min(dataSize, s.bufSize-t)
	part2 := dataSize - part1
	copy(s.b[t:t+part1], b[:part1])
	copy(s.b[:part2], b[part1:part1+part2])

	(*s.tail) = (t + dataSize) % s.bufSize
	if t+dataSize >= s.bufSize {
		(*s.tailCycle) += 1
	}
	return int(dataSize), nil
}
