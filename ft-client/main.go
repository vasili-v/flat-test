package main

import (
	"io"
	"sync"
	"time"

	"github.com/google/flatbuffers/go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/vasili-v/flat-test/echo"
)

func main() {
	pairs := newPairs(total, size)

	c, err := grpc.Dial(server,
		grpc.WithInsecure(),
		grpc.WithCodec(flatbuffers.FlatbuffersCodec{}),
	)
	check(err)
	defer c.Close()

	client := echo.NewEchoClient(c)
	streams := make([]echo.Echo_EchoClient, clients)
	for i := range streams {
		stream, err := client.Echo(context.Background())
		check(err)

		streams[i] = stream
	}

	step := len(pairs) / clients
	if step < 1 {
		step = 1
	}

	var wg sync.WaitGroup
	for i, s := range streams {
		start := i * step
		if start >= len(pairs) {
			break
		}

		end := (i + 1) * step
		if end > len(pairs) || i + 1 >= len(streams) && end < len(pairs) {
			end = len(pairs)
		}

		wg.Add(1)
		go func(s echo.Echo_EchoClient, pairs []*pair) {
			for _, p := range pairs {
				p.sent = time.Now()
				err := s.Send(p.b)
				check(err)

				_, err = s.Recv()
				check(err)

				now := time.Now()
				p.recv = &now
			}

			defer wg.Done()
			err := s.CloseSend()
			check(err)

			var m interface{}
			err = s.RecvMsg(&m)
			if err != io.EOF {
				check(err)
			}
		}(s, pairs[start:end])
	}

	wg.Wait()

	dump(pairs, "")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
