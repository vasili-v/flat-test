package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/google/flatbuffers/go"

	"github.com/vasili-v/flat-test/echo"
)

type pair struct {
	b *flatbuffers.Builder

	sent time.Time
	recv *time.Time
}

func newPairs(n, size int) []*pair {
	out := make([]*pair, n)

	if size > 0 {
		fmt.Fprintf(os.Stderr, "making messages to send:\n")
	}

	for i := range out {
		if size > 0 {
			b := flatbuffers.NewBuilder(0)

			ids := make([]string, size)
			values := make([]string, len(ids))
			for j := range ids {
				ids[j] = randomString(10)
				values[j] = randomString(32)
			}

			if i < 3 {
				fmt.Fprintf(os.Stderr, "\t%d:\n", i)
				for j, id := range ids {
					fmt.Fprintf(os.Stderr,
						"\t\t%q.(%s): %q\n",
						id, echo.EnumNamesType[echo.TypeString], values[j])
				}
				if i < 2 {
					fmt.Fprint(os.Stderr, "\n")
				}
			}

			idOffs := make([]flatbuffers.UOffsetT, len(ids))
			valueOffs := make([]flatbuffers.UOffsetT, len(values))
			for j := range idOffs {
				idOffs[j] = b.CreateString(ids[j])
				valueOffs[j] = b.CreateString(values[j])
			}

			attrs := make([]flatbuffers.UOffsetT, len(idOffs))
			for j := range attrs {
				echo.AttributeStart(b)
				echo.AttributeAddId(b, idOffs[j])
				echo.AttributeAddType(b, echo.TypeString)
				echo.AttributeAddValue(b, valueOffs[j])
				attrs[j] = echo.AttributeEnd(b)
			}

			echo.RequestStartAttributesVector(b, len(attrs))
			for j := len(attrs) - 1; j >= 0; j-- {
				b.PrependUOffsetT(attrs[j])
			}
			attrVec := b.EndVector(len(attrs))

			echo.RequestStart(b)
			echo.RequestAddAttributes(b, attrVec)
			req := echo.RequestEnd(b)
			b.Finish(req)

			out[i] = &pair{b: b}
		} else {
			out[i] = &pair{}
		}
	}

	return out
}

func randomString(n int) string {
	size := rand.Intn(n) + 1

	out := make([]byte, size)
	for i := range out {
		out[i] = byte(0x61 + rand.Intn(0x7B-0x61))
	}

	return string(out)
}
