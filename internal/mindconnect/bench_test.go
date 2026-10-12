package mindconnect

import (
	"slices"
	"testing"
)

func BenchmarkSplitBatches(b *testing.B) {
	reqs := screenRequests()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := SplitBatches(reqs, SafePacket); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeBatch(b *testing.B) {
	groups, err := SplitBatches(screenRequests(), SafePacket)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		encodeGroups(b, groups)
	}
}

func BenchmarkFrameEncode(b *testing.B) {
	reqs := screenRequests()
	b.ReportAllocs()
	for b.Loop() {
		groups, err := SplitBatches(reqs, SafePacket)
		if err != nil {
			b.Fatal(err)
		}
		encodeGroups(b, groups)
	}
}

func BenchmarkFrameValidate(b *testing.B) {
	reqs := screenRequests()
	b.ReportAllocs()
	for b.Loop() {
		for _, r := range reqs {
			if err := r.Validate(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkFrameResponse(b *testing.B) {
	const values = screenW * screenH
	resp := responseOf(values, func(i int) string { return string(rune('a' + i%26)) })
	b.SetBytes(int64(len(resp)))
	b.ReportAllocs()
	for b.Loop() {
		r := NewResponseReader(1, values+1)
		for part := range slices.Chunk(resp, 1460) {
			if _, err := r.Feed(part); err != nil {
				b.Fatal(err)
			}
		}
		if _, _, err := r.Response(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExchangeValues(b *testing.B) {
	reqs := screenRequests()
	groups, err := SplitBatches(reqs, SafePacket)
	if err != nil {
		b.Fatal(err)
	}
	resps := batchResponses(groups)
	b.ReportAllocs()
	for b.Loop() {
		values := make([]Value, 0, len(reqs))
		for i, g := range groups {
			r := NewResponseReader(NextID(i), len(g)+1)
			for part := range slices.Chunk(resps[i], 1460) {
				if _, err := r.Feed(part); err != nil {
					b.Fatal(err)
				}
			}
			if values, err = r.AppendValues(values); err != nil {
				b.Fatal(err)
			}
		}
		if len(values) != len(reqs) {
			b.Fatalf("собрано %d значений, ожидалось %d", len(values), len(reqs))
		}
	}
}
