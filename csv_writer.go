package struct2csv

import (
	"encoding/csv"
	"io"
	"reflect"
	"unsafe"
)

// CSVWriter writes CSV data.
type CSVWriter struct {
	*csv.Writer
	recordCache []string
}

// NewCSVWriter returns new CSVWriter
func NewCSVWriter(w io.Writer) *CSVWriter {
	return &CSVWriter{
		Writer:      csv.NewWriter(w),
		recordCache: make([]string, 0, 16000),
	}
}

// WriteMapping write the csv head mapping file
func (w *CSVWriter) WriteMapping(kvs *KVs) error {
	mapping := kvs.GetMapping()
	header := kvs.GetUnEncodedSortHeader()
	values := make([]string, 0, len(mapping))
	for _, h := range header {
		values = append(values, mapping[h].String())
	}

	return w.WriteAll([][]string{header, values})
}

// WriteCSV writes CSV data.
func (w *CSVWriter) WriteCSV(results *KVs) error {
	header := results.GetEncodedSortHeader()

	if err := w.Write(header); err != nil {
		return err
	}

	// the oriHeader is equal to header but the type
	// why not use header direct? because header is just a string
	// the oriHeader is the type compatibility with results.kvs
	oriHeader := results.GetSortMappingValues()
	for _, result := range results.kvs {
		if result.Len() > 0 { // Kv might have no data because it allocated memory ahead of time
			record := w.toRecord(result, oriHeader)
			if err := w.Write(record); err != nil {
				return err
			}
			w.reset()
		}
	}
	w.Flush()

	return w.Error()
}

func (w *CSVWriter) reset() {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&w.recordCache))
	h.Len = 0
}

func (w *CSVWriter) toRecord(kv *KeyValue, header []KeyType) []string {
	for _, key := range header {
		if value, ok := kv.Get(key); ok {
			w.recordCache = append(w.recordCache, toString(value))
		} else {
			w.recordCache = append(w.recordCache, "")
		}
	}
	return w.recordCache
}
