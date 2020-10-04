package labeler

import "strings"

type keyvalue struct {
	Key   string
	Value string
}

type keyvalueSet struct {
	set      map[string]*keyvalue
	lcaseSet map[string]*keyvalue
	m        map[string]string
}

func newKeyvalueSet(m map[string]string) {
	kvs := keyvalueSet{
		set:      make(map[string]*keyvalue),
		lcaseSet: make(map[string]*keyvalue),
		m:        make(map[string]string),
	}
	kvs.Add(m)
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
	kvs.m[key] = v
}

func (kvs *keyvalueSet) Map() map[string]string {
	return kvs.m
}

func (kvs *keyvalueSet) Delete(key string) {
	delete(kvs.m, key)
	delete(kvs.set, key)
	delete(kvs.lcaseSet, strings.ToLower(key))
}

func (kvs *keyvalueSet) Add(m map[string]string) {
	for k, v := range m {
		kvs.Set(k, v)
	}
}
func (kvs *keyvalueSet) AddSet(v keyvalueSet) {
	kvs.Add(kvs.Map())
}
