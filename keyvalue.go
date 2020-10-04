package labeler

import "strings"

type keyvalue struct {
	Key   string
	Value string
}

type keyvalueSet struct {
	set      map[string]*keyvalue
	lcaseSet map[string]*keyvalue
	vMap     map[string]string // to reduce cycling over set
}

func newKeyValueSet(m map[string]string) {
	kvs := keyvalueSet{
		set:      make(map[string]*keyvalue),
		lcaseSet: make(map[string]*keyvalue),
		vMap:     make(map[string]string),
	}
	for k, v := range m {
		kvs.Set(k, v)
	}
}

func (kvs *keyvalueSet) Get(key string, ignorecase bool) (keyvalue, bool) {
	var kv *keyvalue
	var ok bool
	if ignorecase {
		kv, ok = kvs.lcaseSet[strings.ToLower(key)]
	} else {
		kv, ok = kvs.set[key]
	}
	return *kv, ok
}

func (kvs *keyvalueSet) Set(key string, v string) {
	kv := &keyvalue{Key: key, Value: v}
	kvs.set[key] = kv
	kvs.lcaseSet[strings.ToLower(key)] = kv
	kvs.vMap[key] = v
}

func (kvs *keyvalueSet) GetMap() map[string]string {
	return kvs.vMap
}

func (kvs *keyvalueSet) DeleteFromMap(key string) {
	// delete(kvs.set, key)
	// delete(kvs.lcaseSet, strings.ToLower(key))
	delete(kvs.vMap, key)
}
