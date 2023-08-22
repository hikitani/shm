package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hikitani/shm/bstream"
	"github.com/hikitani/shm/shmem"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
	)
	defer stop()

	shm, err := shmem.Get(777, 32*1024*1024)
	if err != nil {
		log.Fatalf("Failed to get shared memory: %s", err)
	}
	defer shm.Release()

	stream, err := bstream.New(shm.Data())
	if err != nil {
		log.Fatalf("Cannot create stream: %s", err)
	}

	w, err := stream.Writer()
	if err != nil {
		log.Fatalf("Cannot get stream writer: %s", err)
	}
	defer w.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		producer(ctx, w)
	}()

	log.Print("App started")

	<-ctx.Done()
	wg.Wait()

	log.Print("Bye:)")
}

func producer(ctx context.Context, w io.Writer) {
	chunk := make([]byte, 0, 1024)
	t := time.NewTicker(time.Millisecond)

	var i int
	for {
		select {
		case <-t.C:
			chunk = chunk[:0]
			chunk = append(chunk, []byte(fmt.Sprintf("%s: Got Message! Current iteration - %d\n", time.Now().String(), i))...)
			i++

			_, err := w.Write(chunk)
			if err == nil {
				continue
			}
			if errors.Is(err, bstream.ErrNotEnoughMemory) {
				log.Printf("WARN: %s", err)
				continue
			}

			log.Printf("ERR: %s", err)
		case <-ctx.Done():
			return
		}
	}
}
