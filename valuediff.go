package valuediff

import (
	"reflect"
	"sort"
	"strconv"
	"unsafe"
)

type visit struct {
	a1  uintptr
	a2  uintptr
	typ reflect.Type
}

type Diff struct {
	Stack []string
	Src   interface{}
	Dist  interface{}
}

type valueDiff struct {
	visited map[visit]struct{}
	diffs   []Diff
	stack   []string
}

func DeepDiff(x, y interface{}) []Diff {
	v1 := reflect.ValueOf(x)
	v2 := reflect.ValueOf(y)
	return DeepDiffValue(v1, v2)
}

func DeepDiffValue(v1, v2 reflect.Value) []Diff {
	v := valueDiff{
		stack:   []string{},
		visited: map[visit]struct{}{},
	}
	v.deepValueDiff(v1, v2, "")
	return v.diffs
}

func (v *valueDiff) saveDiff(v1, v2 reflect.Value) {
	diff := Diff{
		Stack: cloneStrings(v.stack),
	}
	if v1.Kind() == reflect.Invalid {
		diff.Src = nil
	} else if v1.CanInterface() {
		diff.Src = v1.Interface()
	} else {
		diff.Src = v1.String()
	}

	if v2.Kind() == reflect.Invalid {
		diff.Dist = nil
	} else if v2.CanInterface() {
		diff.Dist = v2.Interface()
	} else {
		diff.Dist = v2.String()
	}
	v.diffs = append(v.diffs, diff)
}

func cloneStrings(s []string) []string {
	n := make([]string, len(s))
	copy(n, s)
	return n
}

func (v *valueDiff) deepValueDiff(v1, v2 reflect.Value, info string) (diff bool) {
	if info != "" {
		v.stack = append(v.stack, info)
		defer func() {
			v.stack = v.stack[:len(v.stack)-1]
		}()
	}
	defer func() {
		if !diff {
			v.saveDiff(v1, v2)
			diff = true
		}
	}()

	if !v1.IsValid() || !v2.IsValid() {
		return v1.IsValid() == v2.IsValid()
	}
	if v1.Type() != v2.Type() {
		return false
	}

	hard := func(v1, v2 reflect.Value) bool {
		switch v1.Kind() {
		case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface:
			return !v1.IsNil() && !v2.IsNil()
		}
		return false
	}

	if hard(v1, v2) {
		addr1 := (*ptrHeader)(unsafe.Pointer(&v1)).ptr
		addr2 := (*ptrHeader)(unsafe.Pointer(&v2)).ptr
		if addr1 > addr2 {
			addr1, addr2 = addr2, addr1
		}

		typ := v1.Type()
		v1 := visit{addr1, addr2, typ}
		if _, ok := v.visited[v1]; ok {
			return true
		}
		v.visited[v1] = struct{}{}
	}

	switch v1.Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len(); i++ {
			v.deepValueDiff(v1.Index(i), v2.Index(i), strconv.FormatInt(int64(i), 10))
		}
		return true
	case reflect.Slice:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		if v1.Len() != v2.Len() {
			return false
		}
		for i := 0; i < v1.Len(); i++ {
			v.deepValueDiff(v1.Index(i), v2.Index(i), strconv.FormatInt(int64(i), 10))
		}
		return true
	case reflect.Interface:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		v.deepValueDiff(v1.Elem(), v2.Elem(), "")
		return true
	case reflect.Ptr:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		v.deepValueDiff(v1.Elem(), v2.Elem(), "")
		return true
	case reflect.Struct:
		t := v1.Type()
		for i, n := 0, v1.NumField(); i < n; i++ {
			v.deepValueDiff(v1.Field(i), v2.Field(i), t.Field(i).Name)
		}
		return true
	case reflect.Map:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		vv1, union, vv2 := unionKey(v1.Type().Key(), v1, v2)

		for _, key := range vv1 {
			v.deepValueDiff(v1.MapIndex(key), reflect.Value{}, key.String())
		}
		for _, key := range vv2 {
			v.deepValueDiff(reflect.Value{}, v2.MapIndex(key), key.String())
		}
		for _, key := range union {
			v.deepValueDiff(v1.MapIndex(key), v2.MapIndex(key), key.String())
		}
		return true
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		if v1.IsNil() || v2.IsNil() {
			return v1.IsNil() == v2.IsNil()
		}
		if v1.Pointer() == v2.Pointer() {
			return true
		}
		fallthrough
	default:
		if v1.CanInterface() && v1.CanInterface() {
			return reflect.DeepEqual(v1.Interface(), v2.Interface())
		}

		return false
	}
}

type ptrHeader struct {
	_   uintptr
	ptr uintptr
}

func unionKey(keyType reflect.Type, m1, m2 reflect.Value) (vv1, union, vv2 []reflect.Value) {
	for iter := m1.MapRange(); iter.Next(); {
		key := iter.Key()
		if m2.MapIndex(key).Kind() == reflect.Invalid {
			vv1 = append(vv1, key)
		} else {
			union = append(union, key)
		}
	}

	for iter := m2.MapRange(); iter.Next(); {
		key := iter.Key()
		if m1.MapIndex(key).Kind() == reflect.Invalid {
			vv2 = append(vv2, key)
		}
	}

	sort.Slice(vv1, func(i, j int) bool {
		return vv1[i].String() < vv1[j].String()
	})
	sort.Slice(union, func(i, j int) bool {
		return union[i].String() < union[j].String()
	})
	sort.Slice(vv2, func(i, j int) bool {
		return vv2[i].String() < vv2[j].String()
	})
	return
}
