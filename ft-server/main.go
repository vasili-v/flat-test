package main

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/google/flatbuffers/go"
	"google.golang.org/grpc"

	"github.com/vasili-v/flat-test/echo"
)

type attribute struct {
	id string
	t  int8
	v  string
}

func newAttributes(in *echo.Request) []attribute {
	size := in.AttributesLength()
	out := make([]attribute, size)
	for i := range out {
		attr := new(echo.Attribute)
		in.Attributes(attr, i)

		out[i] = attribute{
			id: string(attr.Id()),
			t:  attr.Type(),
			v:  string(attr.Value()),
		}
	}

	return out
}

func newResponseAttributes(b *flatbuffers.Builder, in []attribute) flatbuffers.UOffsetT {
	if len(in) <= 0 {
		return 0
	}

	ids := make([]flatbuffers.UOffsetT, len(in))
	values := make([]flatbuffers.UOffsetT, len(in))

	for i, attr := range in {
		ids[i] = b.CreateString(attr.id)
		values[i] = b.CreateString(attr.v)
	}

	attrs := make([]flatbuffers.UOffsetT, len(in))
	for i := range attrs {
		echo.AttributeStart(b)
		echo.AttributeAddId(b, ids[i])
		echo.AttributeAddType(b, in[i].t)
		echo.AttributeAddValue(b, values[i])
		attrs[i] = echo.AttributeEnd(b)
	}

	echo.ResponseStartAttributesVector(b, len(attrs))
	for i := len(attrs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(attrs[i])
	}

	return b.EndVector(len(attrs))
}

func handler(req *echo.Request) *flatbuffers.Builder {
	b := flatbuffers.NewBuilder(0)

	reqAttrs := newAttributes(req)
	resAttrs := newResponseAttributes(b, reqAttrs)

	msg := b.CreateString("Ok")

	echo.ResponseStart(b)
	echo.ResponseAddEffect(b, echo.EffectPong)
	echo.ResponseAddMsg(b, msg)
	if len(reqAttrs) > 0 {
		echo.ResponseAddAttributes(b, resAttrs)
	}
	res := echo.ResponseEnd(b)
	b.Finish(res)

	return b
}

var autoincrement uint64

type server struct{}

func (s *server) Echo(stream echo.Echo_EchoServer) error {
	idx := atomic.AddUint64(&autoincrement, 1)
	fmt.Printf("stream %d started\n", idx)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("stream %d receiving error: %s\n", idx, err)
			return err
		}

		err = stream.Send(handler(req))
		if err != nil {
			fmt.Printf("stream %d sending error: %s\n", idx, err)
			return err
		}
	}

	fmt.Printf("stream %d depleted\n", idx)
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	ln, err := net.Listen("tcp", address)
	check(err)

	opts := []grpc.ServerOption{
		grpc.CustomCodec(flatbuffers.FlatbuffersCodec{}),
	}

	if limit > 0 {
		opts = append(opts,
			grpc.MaxConcurrentStreams(uint32(limit)),
		)
	}

	s := grpc.NewServer(opts...)
	echo.RegisterEchoServer(s, &server{})
	err = s.Serve(ln)
	check(err)
}
