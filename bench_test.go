package optional

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// Sinks keep the compiler from optimising benchmarked calls away.
var (
	sinkInt   int
	sinkBool  bool
	sinkPtr   *int
	sinkBytes []byte
	sinkAny   any
	errSink   error
	sinkVal   Val[int]
)

type benchPayload struct {
	A Val[int]    `json:"a" yaml:"a"`
	B Val[string] `json:"b" yaml:"b"`
	C Val[[]int]  `json:"c" yaml:"c"`
	D Val[int]    `json:"d" yaml:"d"`
}

// benchPointerPayload is the hand-written alternative to Val, for comparison.
type benchPointerPayload struct {
	A *int    `json:"a" yaml:"a"`
	B *string `json:"b" yaml:"b"`
	C *[]int  `json:"c" yaml:"c"`
	D *int    `json:"d" yaml:"d"`
}

var (
	benchValFull = benchPayload{
		A: New(42),
		B: New("hello"),
		C: New([]int{1, 2, 3}),
		D: Empty[int](),
	}
	benchValEmpty = benchPayload{}

	benchInt         = 42
	benchStr         = "hello"
	benchSlice       = []int{1, 2, 3}
	benchPointerFull = benchPointerPayload{A: &benchInt, B: &benchStr, C: &benchSlice}

	benchJSONFull  = []byte(`{"a":42,"b":"hello","c":[1,2,3],"d":null}`)
	benchJSONNulls = []byte(`{"a":null,"b":null,"c":null,"d":null}`)
	benchYAMLFull  = []byte("a: 42\nb: hello\nc:\n    - 1\n    - 2\n    - 3\nd: null\n")
)

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	for i := range b.N {
		sinkVal = New(i)
	}
}

func BenchmarkGet(b *testing.B) {
	b.ReportAllocs()
	v := New(42)

	for range b.N {
		sinkInt, sinkBool = v.Get()
	}
}

func BenchmarkValDefault(b *testing.B) {
	b.ReportAllocs()
	full, empty := New(42), Empty[int]()

	b.Run("present", func(b *testing.B) {
		for range b.N {
			sinkInt = full.ValDefault(1)
		}
	})
	b.Run("empty", func(b *testing.B) {
		for range b.N {
			sinkInt = empty.ValDefault(1)
		}
	})
}

func BenchmarkAsPointer(b *testing.B) {
	b.ReportAllocs()
	v := New(42)

	for range b.N {
		sinkPtr = v.AsPointer() // allocates: the pointer refers to a copy
	}
}

func BenchmarkSetReset(b *testing.B) {
	b.ReportAllocs()
	var v Val[int]

	for i := range b.N {
		v.Set(i)
		v.Reset()
	}
	sinkVal = v
}

func BenchmarkJSONMarshal(b *testing.B) {
	b.Run("val", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			sinkBytes, errSink = json.Marshal(benchValFull)
		}
	})
	b.Run("val_all_empty", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			sinkBytes, errSink = json.Marshal(benchValEmpty)
		}
	})
	b.Run("pointer_baseline", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			sinkBytes, errSink = json.Marshal(benchPointerFull)
		}
	})
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	b.Run("val", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var p benchPayload
			errSink = json.Unmarshal(benchJSONFull, &p)
		}
	})
	b.Run("val_all_null", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var p benchPayload
			errSink = json.Unmarshal(benchJSONNulls, &p)
		}
	})
	b.Run("pointer_baseline", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var p benchPointerPayload
			errSink = json.Unmarshal(benchJSONFull, &p)
		}
	})
}

func BenchmarkYAMLMarshal(b *testing.B) {
	b.Run("val", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			sinkBytes, errSink = yaml.Marshal(benchValFull)
		}
	})
	b.Run("pointer_baseline", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			sinkBytes, errSink = yaml.Marshal(benchPointerFull)
		}
	})
}

func BenchmarkYAMLUnmarshal(b *testing.B) {
	b.Run("val", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var p benchPayload
			errSink = yaml.Unmarshal(benchYAMLFull, &p)
		}
	})
	b.Run("pointer_baseline", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			var p benchPointerPayload
			errSink = yaml.Unmarshal(benchYAMLFull, &p)
		}
	})
}

func BenchmarkScan(b *testing.B) {
	b.Run("int64_exact", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[int64]
		for range b.N {
			errSink = v.Scan(int64(42))
		}
	})
	b.Run("int_converted", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[int]
		for range b.N {
			errSink = v.Scan(int64(42))
		}
	})
	b.Run("string_from_bytes", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[string]
		src := []byte("hello")
		for range b.N {
			errSink = v.Scan(src)
		}
	})
	b.Run("bytes_copied", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[[]byte]
		src := []byte("hello")
		for range b.N {
			errSink = v.Scan(src)
		}
	})
	b.Run("null", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[int64]
		for range b.N {
			errSink = v.Scan(nil)
		}
	})
	b.Run("scanner_type", func(b *testing.B) {
		b.ReportAllocs()
		var v Val[sql.NullBool]
		for range b.N {
			errSink = v.Scan(true)
		}
	})
	b.Run("sql_Null_baseline", func(b *testing.B) {
		b.ReportAllocs()
		var v sql.Null[int64]
		for range b.N {
			errSink = v.Scan(int64(42))
		}
	})
}

func BenchmarkValue(b *testing.B) {
	var res driver.Value

	b.Run("int64_exact", func(b *testing.B) {
		b.ReportAllocs()
		v := New[int64](42)
		for range b.N {
			res, errSink = v.Value()
		}
	})
	b.Run("int_converted", func(b *testing.B) {
		b.ReportAllocs()
		v := New(42)
		for range b.N {
			res, errSink = v.Value()
		}
	})
	b.Run("string", func(b *testing.B) {
		b.ReportAllocs()
		v := New("hello")
		for range b.N {
			res, errSink = v.Value()
		}
	})
	b.Run("valuer_type", func(b *testing.B) {
		b.ReportAllocs()
		v := New(sql.NullBool{Bool: true, Valid: true})
		for range b.N {
			res, errSink = v.Value()
		}
	})
	b.Run("empty", func(b *testing.B) {
		b.ReportAllocs()
		v := Empty[int64]()
		for range b.N {
			res, errSink = v.Value()
		}
	})

	sinkAny = res
}
