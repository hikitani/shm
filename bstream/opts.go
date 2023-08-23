package bstream

type writerOption func(s *Writer) error

func OverwriteOldData() writerOption {
	return func(w *Writer) error {
		w.overwriteOldData = true
		return nil
	}
}
