package tags

import "sort"

// Tags holds arbitrary key-value metadata attached to a monitor entry.
type Tags map[string]string

// New returns an empty Tags map.
func New() Tags {
	return make(Tags)
}

// Set adds or updates a key-value pair.
func (t Tags) Set(key, value string) {
	t[key] = value
}

// Get returns the value for a key and whether it was found.
func (t Tags) Get(key string) (string, bool) {
	v, ok := t[key]
	return v, ok
}

// Delete removes a key from the tags.
func (t Tags) Delete(key string) {
	delete(t, key)
}

// Keys returns a sorted list of all tag keys.
func (t Tags) Keys() []string {
	keys := make([]string, 0, len(t))
	for k := range t {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Clone returns a shallow copy of the tags.
func (t Tags) Clone() Tags {
	copy := make(Tags, len(t))
	for k, v := range t {
		copy[k] = v
	}
	return copy
}

// Matches reports whether all pairs in filter are present and equal in t.
func (t Tags) Matches(filter Tags) bool {
	for k, v := range filter {
		if t[k] != v {
			return false
		}
	}
	return true
}
