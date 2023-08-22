package bstream

type Reader struct {
	*Stream
}

func (s *Reader) Close() error {
	s.rcancel()
	return nil
}

func (s *Reader) Read(b []byte) (int, error) {
	select {
	case <-s.rctx.Done():
		return 0, ErrReaderClosed
	default:
	}

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
	used := s.bufSize - availSize
	if used == 0 {
		return 0, ErrStreamEmpty
	}
	dataSize := min(uint64(len(b)), used)

	part1 := min(dataSize, s.bufSize-h)
	part2 := dataSize - part1
	copy(b[:part1], s.b[h:h+part1])
	copy(b[part1:part1+part2], b[:part2])

	(*s.head) = (h + dataSize) % s.bufSize
	if h+dataSize >= s.bufSize {
		(*s.headCycle) += 1
	}
	return int(dataSize), nil
}
