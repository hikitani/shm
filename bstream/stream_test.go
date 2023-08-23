package bstream

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamNormal(t *testing.T) {
	var buf [56 + 13 + 1]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	w, err := s.Writer()
	require.NoError(t, err)
	require.NotNil(t, w)
	r, err := s.Reader()
	require.NoError(t, err)
	require.NotNil(t, r)

	data := [][]byte{
		[]byte("hello"),
		[]byte("foo"),
		[]byte("ok"),
		[]byte("bar"),
	}

	for _, msg := range data {
		n, err := w.Write(msg)
		assert.NoError(t, err)
		assert.Equal(t, len(msg), n)
	}

	for _, expected := range data {
		actual := make([]byte, len(expected))
		n, err := r.Read(actual)
		assert.NoError(t, err)
		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, actual)
	}
}

func TestStreamIsFull(t *testing.T) {
	var buf [56 + 5 + 1]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	w, err := s.Writer()
	require.NoError(t, err)
	require.NotNil(t, w)

	n, err := w.Write([]byte("hello"))
	require.Equal(t, 5, n)
	require.NoError(t, err)
	n, err = w.Write([]byte("hello"))
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, ErrNotEnoughMemory)
}

func TestStreamIsEmpty(t *testing.T) {
	var buf [56 + 5 + 1]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	r, err := s.Reader()
	require.NoError(t, err)
	require.NotNil(t, r)

	b := make([]byte, 5)
	n, err := r.Read(b)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, ErrStreamEmpty)
}

func TestStreamIsClosed(t *testing.T) {
	var buf [56 + 5 + 1]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	w, err := s.Writer()
	require.NoError(t, err)
	require.NotNil(t, w)
	r, err := s.Reader()
	require.NoError(t, err)
	require.NotNil(t, r)

	require.NoError(t, w.Close())

	n, err := w.Write([]byte("hello"))
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, ErrWriterClosed)

	require.NoError(t, r.Close())

	b := make([]byte, 5)
	n, err = r.Read(b)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, ErrReaderClosed)
}

func TestStreamProducerBusy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var buf [56]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	w, err := s.Writer()
	require.NoError(t, err)
	require.NotNil(t, w)

	w, err = s.Writer()
	assert.ErrorIs(t, err, ErrStreamWriterBusy)
	assert.Nil(t, w)
}

func TestStreamAcquireProducer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var buf [56]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	w, err := s.Writer()
	require.NoError(t, err)
	require.NotNil(t, w)

	require.NoError(t, w.Close())

	w, err = s.Writer()
	assert.NoError(t, err)
	assert.NotNil(t, w)
}

func TestStreamConsumerBusy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var buf [56]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	r, err := s.Reader()
	require.NoError(t, err)
	require.NotNil(t, r)

	r, err = s.Reader()
	assert.ErrorIs(t, err, ErrStreamReaderBusy)
	assert.Nil(t, r)
}

func TestStreamAcquireConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var buf [56]byte
	s, err := New(buf[:])
	require.NoError(t, err)
	require.NotNil(t, s)
	r, err := s.Reader()
	require.NoError(t, err)
	require.NotNil(t, r)

	require.NoError(t, r.Close())

	r, err = s.Reader()
	assert.NoError(t, err)
	assert.NotNil(t, r)
}
