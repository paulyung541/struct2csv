package struct2csv

import (
	"testing"
)

// BenchmarkPathBuilder_Clone-8   	13933215	        83.01 ns/op	     224 B/op	       3 allocs/op
func BenchmarkPathBuilder_Clone(b *testing.B) {
	b.ReportAllocs()
	jp := NewPathBuilder(0)
	jp.AppendString("/properties/apiData/0/payload/satellite_feature_data/histograms/0/histogram/bin_values/0")

	for i := 0; i < b.N; i++ {
		jp.Clone(0)
	}
}
