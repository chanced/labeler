package labeler

import "strings"

type keyvalue struct {
	Key   string
	Value string
}

type keyvalues struct {
	lookup map[string]*keyvalue
	lcase  map[string]*keyvalue
	m      map[string]string
}

func newKeyvalues() keyvalues {
	kvs := keyvalues{
		lookup: make(map[string]*keyvalue),
		lcase:  make(map[string]*keyvalue),
		m:      make(map[string]string),
	}
	return kvs
}

func (kvs *keyvalues) Get(key string, ignorecase bool) (keyvalue, bool) {
	var kv *keyvalue
	var ok bool
	if ignorecase {
		kv, ok = kvs.lcase[strings.ToLower(key)]
	} else {
		kv, ok = kvs.lookup[key]
	}
	if ok {
		return *kv, ok
	}
	return keyvalue{}, ok

}

func (kvs *keyvalues) Set(key string, v string) {
	kv := &keyvalue{Key: key, Value: v}
	kvs.lookup[key] = kv
	kvs.lcase[strings.ToLower(key)] = kv
	kvs.m[key] = v
}

func (kvs *keyvalues) Map() map[string]string {
	return kvs.m
}

func (kvs *keyvalues) Delete(key string) {
	delete(kvs.m, key)
	delete(kvs.lookup, key)
	delete(kvs.lcase, strings.ToLower(key))
}

func (kvs *keyvalues) Add(m map[string]string) {
	if m == nil {
		return
	}
	for k, v := range m {
		kvs.Set(k, v)
	}
}
func (kvs *keyvalues) AddSet(v keyvalues) {
	kvs.Add(kvs.Map())
}
