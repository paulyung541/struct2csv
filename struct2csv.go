package struct2csv

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// HeaderConverter will convert the path string of the original struct field
// to the custom of KeyType
type HeaderConverter interface {
	ConvertHeader(string) KeyType
	Reset()
}

// compile time checks
var _ HeaderConverter = (*HeaderAutoIncrementConv)(nil)
var _ HeaderConverter = (*HeaderOriginalStringConv)(nil)

type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
	opts := defaultOpts()
	for _, option := range options {
		option(opts)
	}

	return opts
}

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

type Options struct {
	resultCap     int  // pre-allocated for []KeyValue
	isObjArray    bool // if u know the input data is must the map|struct of slice, set it true
	strBuilderCap int  // pre-allocated for strings.Builder Cap size, call the Grow function
	rowSize       int  // pre-allocated for KeyValue map size
}

func WithResultCap(p int) Option {
	return func(opts *Options) {
		opts.resultCap = p
	}
}

func WithIsObjArray(p bool) Option {
	return func(opts *Options) {
		opts.isObjArray = p
	}
}

func WithStrBuilderCap(p int) Option {
	return func(opts *Options) {
		opts.strBuilderCap = p
	}
}

func WithRowSize(p int) Option {
	return func(opts *Options) {
		opts.rowSize = p
	}
}

func defaultOpts() *Options {
	return &Options{
		resultCap:     50,
		isObjArray:    true,
		strBuilderCap: 100,
		rowSize:       18000,
	}
}

type StructConverter struct {
	kvs        *KVs
	opts       *Options
	headerConv HeaderConverter
}

// NewStructConverter a converter can convert struct to csv kv
func NewStructConverter(headerConv HeaderConverter, opts ...Option) (*StructConverter, error) {
	if headerConv == nil {
		return nil, errors.New("HeaderConverter can not be nil")
	}

	sc := &StructConverter{
		opts:       loadOptions(opts...),
		headerConv: headerConv,
	}
	sc.kvs = NewKVs(sc.opts.resultCap, sc.opts.rowSize)

	return sc, nil
}

// Convert converts Struct to CSV key value
func (s *StructConverter) Convert(data interface{}) (*KVs, error) {
	s.headerConv.Reset()

	v := valueOf(data)
	sliceIterator := func() error {
		for i := 0; i < v.Len(); i++ {
			_, err := s.doFlatten(v.Index(i), i)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if s.opts.isObjArray {
		if err := sliceIterator(); err != nil {
			return nil, err
		}
		return s.kvs, nil
	}

	switch v.Kind() {
	case reflect.Map:
		if v.Len() > 0 {
			result, err := s.doFlatten(v, -1)
			if err != nil {
				return nil, err
			}
			s.kvs.appendElem(result)
		}
	case reflect.Slice:
		if isObjectArray(v) {
			if err := sliceIterator(); err != nil {
				return nil, err
			}
		} else if v.Len() > 0 {
			result, err := s.doFlatten(v, -1)
			if err != nil {
				return nil, err
			}
			if result != nil {
				s.kvs.appendElem(result)
			}
		}
	case reflect.Struct:
		result, err := s.doFlatten(v, -1)
		if err != nil {
			return nil, err
		}
		s.kvs.appendElem(result)
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			result, err := s.doFlatten(v, -1)
			if err != nil {
				return nil, err
			}
			s.kvs.appendElem(result)
		}
	default:
		return nil, fmt.Errorf("Unsupported struct structure, kind = %s", v.Kind().String())
	}

	return s.kvs, nil
}

func (s *StructConverter) Clear() {
	s.kvs.Clear()
	s.kvs = nil
}

func isObjectArray(obj interface{}) bool {
	value := valueOf(obj)
	if value.Kind() != reflect.Slice {
		return false
	}

	l := value.Len()
	if l == 0 {
		return false
	}
	kind := valueOf(value.Index(0)).Kind()
	if kind != reflect.Map && kind != reflect.Ptr && kind != reflect.Struct {
		return false
	}

	return true
}

func valueOf(obj interface{}) reflect.Value {
	v, ok := obj.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(obj)
	}

	for v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

func toString(obj interface{}) string {
	switch v := obj.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.Itoa(int(v))
	case int16:
		return strconv.Itoa(int(v))
	case int32:
		return strconv.Itoa(int(v))
	case int64:
		return strconv.Itoa(int(v))
	case uint:
		return strconv.Itoa(int(v))
	case uint8:
		return strconv.Itoa(int(v))
	case uint16:
		return strconv.Itoa(int(v))
	case uint32:
		return strconv.Itoa(int(v))
	case uint64:
		return strconv.Itoa(int(v))
	case float32:
		return strconv.FormatFloat(float64(v), 'f', 10, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', 10, 64)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", obj)
	}
}

func protoMessage(i interface{}) (protoreflect.Value, bool) {
	msg, ok := i.(proto.Message)
	if !ok {
		return protoreflect.Value{}, false
	}

	v := protoreflect.ValueOf(msg.ProtoReflect())
	return v, true
}
