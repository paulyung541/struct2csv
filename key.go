package struct2csv

import (
	"crypto/md5"
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

type KeyType interface {
	String() string // get encode string
	Int() uint64    // get the key map to uint64
}

// compile time checks
var _ KeyType = KeyAutoIncrementID(0)
var _ KeyType = KeyString("")

type KVs struct {
	preSize         int
	preMappingSize  int
	kvs             []*KeyValue
	mapping         map[string]KeyType
	encodeHeaders   []string // mapping's value call there's String()
	unEncodeHeaders []string // mapping's key
}

func NewKVs(size, preMappingSize int) *KVs {
	kvs := &KVs{
		preSize:         size,
		kvs:             make([]*KeyValue, 0, size),
		mapping:         make(map[string]KeyType, preMappingSize),
		encodeHeaders:   make([]string, 0, preMappingSize),
		unEncodeHeaders: make([]string, 0, preMappingSize),
	}

	for i := 0; i < size; i++ {
		kvs.kvs = append(kvs.kvs, newKeyValue(preMappingSize))
	}

	return kvs
}

func (kvs *KVs) Reset() {
	enH := (*reflect.SliceHeader)(unsafe.Pointer(&kvs.encodeHeaders))
	enH.Len = 0
	unH := (*reflect.SliceHeader)(unsafe.Pointer(&kvs.unEncodeHeaders))
	unH.Len = 0
	kvs.mapping = make(map[string]KeyType, kvs.preSize)
}

func (kvs *KVs) Clear() {
	kvs.kvs = nil
	kvs.mapping = nil
	kvs.encodeHeaders = nil
	kvs.unEncodeHeaders = nil
}

func (kvs *KVs) getKVElem(index int) *KeyValue {
	return kvs.kvs[index]
}

func (kvs *KVs) appendElem(kv *KeyValue) {
	kvs.kvs = append(kvs.kvs, kv)
}

func (kvs *KVs) GetMapping() map[string]KeyType {
	return kvs.mapping
}

func (kvs *KVs) GetSortMappingValues() []KeyType {
	vs := make([]KeyType, 0, len(kvs.mapping))
	for _, v := range kvs.mapping {
		vs = append(vs, v)
	}

	sort.SliceStable(vs, func(i, j int) bool {
		return vs[i].String() < vs[j].String()
	})

	return vs
}

func (kvs *KVs) GetUnEncodedSortHeader() []string {
	if len(kvs.unEncodeHeaders) > 0 {
		return kvs.unEncodeHeaders
	}

	for unen := range kvs.mapping {
		kvs.unEncodeHeaders = append(kvs.unEncodeHeaders, unen)
	}
	sort.Strings(kvs.unEncodeHeaders)
	return kvs.unEncodeHeaders
}

func (kvs *KVs) GetEncodedSortHeader() []string {
	if len(kvs.encodeHeaders) > 0 {
		return kvs.encodeHeaders
	}

	for _, en := range kvs.mapping {
		kvs.encodeHeaders = append(kvs.encodeHeaders, en.String())
	}
	sort.Strings(kvs.encodeHeaders)
	return kvs.encodeHeaders
}

func (kvs *KVs) SetEncodedSortHeader(header []string) {
	kvs.encodeHeaders = header
}

type WrapperValue struct {
	isValid bool // have use data
	value   interface{}
}

func (w *WrapperValue) clear() {
	w.isValid = false
}

// KeyValue records the struct's key(containing the path) and its value
// consider the performance problems caused by map keys of the interface type
// uint64 is used as the map key, and wrapped by KeyValue.
// call Set and Get of map's read and write. KeyType can be extended to different types,
// then both performance and code extensibility are considered
type KeyValue struct {
	kv map[uint64]*WrapperValue // all KeyType must be map to a unique uint64, use KeyType's Int() function
}

func newKeyValue(preSize int) *KeyValue {
	return &KeyValue{kv: make(map[uint64]*WrapperValue, preSize)}
}

func (t *KeyValue) Set(k KeyType, v interface{}) {
	t.kv[k.Int()] = &WrapperValue{
		isValid: true,
		value:   v,
	}
}

func (t *KeyValue) Get(k KeyType) (interface{}, bool) {
	v, ok := t.kv[k.Int()]
	if !ok || !v.isValid {
		return nil, false
	}
	v.clear()
	return v.value, true
}

func (t *KeyValue) Len() int {
	return len(t.kv)
}

type KeyAutoIncrementID uint64

func newKeyAuto(max uint64) KeyAutoIncrementID {
	max++
	return KeyAutoIncrementID(max)
}

func (k KeyAutoIncrementID) String() string {
	return strconv.Itoa(int(k))
}

func (k KeyAutoIncrementID) Int() uint64 {
	return uint64(k)
}

type KeyString string

func newKeyString(s string) KeyString {
	return KeyString(s)
}

func (k KeyString) String() string {
	return string(k)
}

func (k KeyString) Int() uint64 {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&k))
	sliceH := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	res := md5.Sum(*(*[]byte)(unsafe.Pointer(sliceH)))
	return *(*uint64)(unsafe.Pointer(&res)) // only use 8 byte
}
