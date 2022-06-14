package struct2csv

import (
	"fmt"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// if index = -1, don't use cache map, caller must use append the result to slice
func (s *StructConverter) doFlatten(obj interface{}, index int) (*KeyValue, error) {
	var f *KeyValue
	if index == -1 {
		f = newKeyValue(s.opts.rowSize)
	} else {
		f = s.kvs.getKVElem(index)
	}
	key := NewPathBuilder(s.opts.strBuilderCap)
	if err := s.flatten(f, obj, key); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *StructConverter) flatten(out *KeyValue, obj interface{}, key PathBuilder) error {
	value, ok := obj.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(obj)
	}
	for value.Kind() == reflect.Interface {
		value = value.Elem()
	}

	if !value.IsValid() || value.IsZero() {
		return nil
	}

	switch value.Kind() {
	case reflect.Map:
		return s.flattenMap(out, value, key)
	case reflect.Slice, reflect.Array:
		return s.flattenSlice(out, value, key)
	case reflect.Struct:
		return s.flattenStruct(out, value, key)
	case reflect.String:
		s.set(out, key.String(), value.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s.set(out, key.String(), value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s.set(out, key.String(), value.Uint())
	case reflect.Float32, reflect.Float64:
		s.set(out, key.String(), value.Float())
	case reflect.Bool:
		s.set(out, key.String(), value.Bool())
	case reflect.Ptr:
		if value.Elem().Kind() == reflect.Struct {
			return s.flattenStruct(out, value, key)
		} else {
			return s.flatten(out, value.Elem(), key)
		}
	default:
		return fmt.Errorf("unknown kind: %s", value.Kind())
	}
	return nil
}

func (s *StructConverter) flattenMap(out *KeyValue, value reflect.Value, prefix PathBuilder) error {
	for _, k := range value.MapKeys() {
		vv := value.MapIndex(k)
		if !vv.IsValid() || vv.IsZero() {
			continue
		}

		pointer := prefix.Clone(s.opts.strBuilderCap)
		pointer.AppendString(k.String())
		if err := s.flatten(out, vv, pointer); err != nil {
			return err
		}
	}

	return nil
}

func (s *StructConverter) flattenSlice(out *KeyValue, value reflect.Value, prefix PathBuilder) error {
	for i := 0; i < value.Len(); i++ {
		vv := value.Index(i)
		if !vv.IsValid() || vv.IsZero() {
			continue
		}

		pointer := prefix.Clone(s.opts.strBuilderCap)
		pointer.AppendString(strconv.Itoa(i))
		if err := s.flatten(out, vv, pointer); err != nil {
			return err
		}
	}

	return nil
}

func (s *StructConverter) flattenStruct(out *KeyValue, value reflect.Value, prefix PathBuilder) error {
	if !value.CanInterface() {
		return nil
	}

	protoValue, ok := protoMessage(value.Interface())

	if !ok {
		// normal struct
		tp := value.Type()
		if tp.Kind() == reflect.Ptr {
			tp = tp.Elem()
			value = value.Elem()
		}

		for i := 0; i < tp.NumField(); i++ {
			f := tp.Field(i)
			vv := value.FieldByName(f.Name)
			if !vv.IsValid() || vv.IsZero() {
				continue
			}

			pointer := prefix.Clone(s.opts.strBuilderCap)
			pointer.AppendString(f.Name)
			if err := s.flatten(out, vv, pointer); err != nil {
				return err
			}
		}
		return nil
	}

	// proto struct
	return s.flattenProtoStruct(out, protoValue, prefix)
}

// resolve proto struct
func (s *StructConverter) flattenProtoStruct(out *KeyValue, value protoreflect.Value, prefix PathBuilder) error {
	msg := value.Message()
	var err error
	msg.Range(func(fd protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		if !value.IsValid() {
			return true
		}

		pointer := prefix.Clone(s.opts.strBuilderCap)
		pointer.AppendString(fd.TextName())
		err = s.flattenProto(out, fd, value, pointer)
		return err == nil
	})

	return err
}

func (s *StructConverter) flattenProto(out *KeyValue, fd protoreflect.FieldDescriptor, value protoreflect.Value, key PathBuilder) error {
	if !value.IsValid() {
		return nil
	}

	normal := func() error {
		switch v := value.Interface().(type) {
		case protoreflect.Message:
			return s.flattenProtoStruct(out, value, key)
		case protoreflect.List:
			return s.flattenProtoSlice(out, value, key)
		case protoreflect.Map:
			return s.flattenProtoMap(out, value, key)
		case protoreflect.Enum, protoreflect.EnumNumber:
			s.set(out, key.String(), int32(value.Enum()))
		case int, int8, int16, int32, int64:
			s.set(out, key.String(), value.Int())
		case uint, uint8, uint16, uint32, uint64:
			s.set(out, key.String(), value.Uint())
		case float32, float64:
			s.set(out, key.String(), value.Float())
		case bool:
			s.set(out, key.String(), value.Bool())
		case string:
			s.set(out, key.String(), value.String())
		default:
			return fmt.Errorf("flattenProto: unknow type %v, value == %#v, keypoint = %v", v, value.Interface(), key.String())
		}

		return nil
	}

	if fd == nil {
		return normal()
	}

	if fd.IsList() {
		return s.flattenProtoSlice(out, value, key)
	}

	if fd.IsMap() {
		return s.flattenProtoMap(out, value, key)
	}

	switch fd.Kind() {
	case protoreflect.BoolKind:
		s.set(out, key.String(), value.Bool())
	case protoreflect.EnumKind:
		s.set(out, key.String(), value.Enum())
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		s.set(out, key.String(), value.Int())
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		s.set(out, key.String(), value.Uint())
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		s.set(out, key.String(), value.Float())
	case protoreflect.StringKind:
		s.set(out, key.String(), value.String())
	case protoreflect.MessageKind:
		return s.flattenProtoStruct(out, value, key)
	}

	return nil
}

func (s *StructConverter) flattenProtoSlice(out *KeyValue, value protoreflect.Value, prefix PathBuilder) error {
	list := value.List()
	for i := 0; i < list.Len(); i++ {
		elem := list.Get(i)

		if !elem.IsValid() {
			continue
		}

		pointer := prefix.Clone(s.opts.strBuilderCap)
		pointer.AppendString(strconv.Itoa(i))
		if err := s.flattenProto(out, nil, elem, pointer); err != nil {
			return err
		}
	}

	return nil
}

func (s *StructConverter) flattenProtoMap(out *KeyValue, value protoreflect.Value, prefix PathBuilder) error {
	var err error
	value.Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		if !v.IsValid() {
			return true
		}

		pointer := prefix.Clone(s.opts.strBuilderCap)
		pointer.AppendString(k.String())
		err = s.flattenProto(out, nil, v, pointer)
		return err == nil
	})

	return err
}

func (s *StructConverter) set(out *KeyValue, k string, v interface{}) {
	kt, ok := s.kvs.mapping[k]
	if !ok {
		kt = s.headerConv.ConvertHeader(k)
	}

	s.kvs.mapping[k] = kt
	out.Set(kt, v)
}
