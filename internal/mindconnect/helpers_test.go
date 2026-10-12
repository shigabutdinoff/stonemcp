package mindconnect

import (
	"strconv"
	"testing"
)

const screenW, screenH = 90, 40

func screenRequests() []Request {
	reqs := make([]Request, 0, screenW*screenH)
	for y := range screenH {
		for x := range screenW {
			reqs = append(reqs, Var("draw.GetSymbol", Int(x), Int(y)))
		}
	}
	return reqs
}

func encodeGroups(tb testing.TB, groups [][]Request) {
	tb.Helper()
	for i, g := range groups {
		if _, _, err := EncodeBatch(NextID(i), g, SafePacket); err != nil {
			tb.Fatal(err)
		}
	}
}

func responseOf(values int, value func(i int) string) []byte {
	resp := []byte("1")
	for i := range values {
		resp = append(resp, FieldSep+value(i)...)
	}
	return append(resp, FieldSep+"True"...)
}

func batchResponses(groups [][]Request) [][]byte {
	resps := make([][]byte, len(groups))
	for i, g := range groups {
		resp := strconv.AppendInt(nil, int64(NextID(i)), 10)
		for range g {
			resp = append(resp, FieldSep+"a"...)
		}
		resps[i] = append(resp, FieldSep+"True"...)
	}
	return resps
}
