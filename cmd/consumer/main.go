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

	r, err := stream.Reader()
	if err != nil {
		log.Fatalf("Cannot get stream reader: %s", err)
	}
	defer r.Close()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer(ctx, r)
	}()

	log.Print("App started")

	<-ctx.Done()
	wg.Wait()

	log.Print("Bye:)")
}

func consumer(ctx context.Context, r io.Reader) {
	chunk := make([]byte, 1024)
	t := time.NewTicker(time.Millisecond)

	var row []byte
	for {
		select {
		case <-t.C:
			n, err := r.Read(chunk)
			if err == nil {
				b := chunk[:n]

				var flushed bool
				for i, el := range b {
					if el == '\n' {
						row = append(row, b[:i+1]...)
						fmt.Print(string(row))
						row = append(row[:0], b[i+1:]...)
						flushed = true
						break
					}
				}

				if !flushed {
					row = append(row, b...)
				}
				continue
			}
			if errors.Is(err, bstream.ErrStreamEmpty) {
				log.Printf("WARN: %s", err)
				continue
			}

			log.Printf("ERR: %s", err)
		case <-ctx.Done():
			return
		}
	}
}
