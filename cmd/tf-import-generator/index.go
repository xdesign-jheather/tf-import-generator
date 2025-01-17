package main

import "sort"

type Index map[string][]string

func (i Index) Add(key, item string) {
	i[key] = append(i[key], item)
}

func (i Index) Keys() []string {
	var keys []string

	for key := range i {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (i Index) Walk(f func(int, string, []string)) {
	for ii, key := range i.Keys() {
		items := make([]string, len(i[key]))

		copy(items, i[key])

		sort.Strings(items)

		f(ii+1, key, items)
	}
}
