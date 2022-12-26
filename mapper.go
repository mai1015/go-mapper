package go_mapper

import (
	"errors"
	"fmt"
	"reflect"
)

type MapperFunc func(any) (any, error)

type IMapper interface {
	Map(source, dest interface{}, loose bool)

	RegisterMapping(from, to string, f MapperFunc)
	UnregisterMapping(from, to string)
}

type defaultMapper struct {
	valueMap map[string]map[string]MapperFunc
}

func NewDefaultMapper() IMapper {
	return &defaultMapper{
		make(map[string]map[string]MapperFunc),
	}
}

func (d *defaultMapper) Map(source, dest interface{}, loose bool) {
	var destType = reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		panic("Dest must be a pointer type")
	}
	var sourceVal = reflect.ValueOf(source)
	var destVal = reflect.ValueOf(dest).Elem()
	d.mapValues(sourceVal, destVal, loose)
}

func (d *defaultMapper) RegisterMapping(from, to string, f MapperFunc) {
	v, ok := d.valueMap[from]
	if !ok {
		v = make(map[string]MapperFunc)
		d.valueMap[from] = v
	}
	v[to] = f
}

func (d *defaultMapper) UnregisterMapping(from, to string) {
	if k, ok := d.valueMap[from]; ok {
		if _, ok := k[to]; ok {
			delete(k, to)
		}
	}
}

func (d *defaultMapper) mapCustom(source, destVal reflect.Value) error {
	s := source.Type().String()
	t := destVal.Type().String()

	g, ok := d.valueMap[s]
	if !ok {
		return errors.New("cannot find convertor")
	}
	f, ok := g[t]
	if !ok {
		return errors.New("cannot find convertor")
	}

	v, err := f(source.Interface())
	if err != nil {
		return err
	}

	destVal.Set(reflect.ValueOf(v))
	return nil
}

func (d *defaultMapper) mapValues(sourceVal, destVal reflect.Value, loose bool) {
	destType := destVal.Type()
	if destType.Kind() == reflect.Struct {
		if sourceVal.Type().Kind() == reflect.Ptr {
			if sourceVal.IsNil() {
				// If source is nil, it maps to an empty struct
				sourceVal = reflect.New(sourceVal.Type().Elem())
			}
			sourceVal = sourceVal.Elem()
		}
		for i := 0; i < destVal.NumField(); i++ {
			d.mapField(sourceVal, destVal, i, loose)
		}
	} else if destType == sourceVal.Type() {
		destVal.Set(sourceVal)
	} else if destType.Kind() == reflect.Ptr {
		if d.valueIsNil(sourceVal) {
			return
		}
		val := reflect.New(destType.Elem())
		d.mapValues(sourceVal, val.Elem(), loose)
		destVal.Set(val)
	} else if destType.Kind() == reflect.Slice {
		d.mapSlice(sourceVal, destVal, loose)
	} else {
		err := d.mapCustom(sourceVal, destVal)
		if err != nil {
			panic("Currently not supported")
		}
	}
}

func (d *defaultMapper) mapSlice(sourceVal, destVal reflect.Value, loose bool) {
	destType := destVal.Type()
	length := sourceVal.Len()
	target := reflect.MakeSlice(destType, length, length)
	for j := 0; j < length; j++ {
		val := reflect.New(destType.Elem()).Elem()
		d.mapValues(sourceVal.Index(j), val, loose)
		target.Index(j).Set(val)
	}

	if length == 0 {
		d.verifyArrayTypesAreCompatible(sourceVal, destVal, loose)
	}
	destVal.Set(target)
}

func (d *defaultMapper) verifyArrayTypesAreCompatible(sourceVal, destVal reflect.Value, loose bool) {
	dummyDest := reflect.New(reflect.PtrTo(destVal.Type()))
	dummySource := reflect.MakeSlice(sourceVal.Type(), 1, 1)
	d.mapValues(dummySource, dummyDest.Elem(), loose)
}

func (d *defaultMapper) mapField(source, destVal reflect.Value, i int, loose bool) {
	destType := destVal.Type()
	fieldName := destType.Field(i).Name
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("Error mapping field: %s. DestType: %v. SourceType: %v. Error: %v", fieldName, destType, source.Type(), r))
		}
	}()

	destField := destVal.Field(i)
	if destType.Field(i).Anonymous {
		d.mapValues(source, destField, loose)
	} else {
		if d.valueIsContainedInNilEmbeddedType(source, fieldName) {
			return
		}
		sourceField := source.FieldByName(fieldName)
		if (sourceField == reflect.Value{}) {
			if loose {
				return
			}
			if destField.Kind() == reflect.Struct {
				d.mapValues(source, destField, loose)
				return
			} else {
				for i := 0; i < source.NumField(); i++ {
					if source.Field(i).Kind() != reflect.Struct {
						continue
					}
					if sourceField = source.Field(i).FieldByName(fieldName); (sourceField != reflect.Value{}) {
						break
					}
				}
			}
		}
		d.mapValues(sourceField, destField, loose)
	}
}

func (d *defaultMapper) valueIsNil(value reflect.Value) bool {
	return value.Type().Kind() == reflect.Ptr && value.IsNil()
}

func (d *defaultMapper) valueIsContainedInNilEmbeddedType(source reflect.Value, fieldName string) bool {
	structField, _ := source.Type().FieldByName(fieldName)
	ix := structField.Index
	if len(structField.Index) > 1 {
		parentField := source.FieldByIndex(ix[:len(ix)-1])
		if d.valueIsNil(parentField) {
			return true
		}
	}
	return false
}
