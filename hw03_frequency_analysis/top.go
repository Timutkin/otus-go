package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type Entry struct {
	Key   string
	Value int
}

func Top10(str string) []string {
	if len(str) == 0 {
		return make([]string, 0)
	}
	words := strings.Fields(str)
	statics := make(map[string]int, len(words))
	for _, v := range words {
		statics[v]++
	}
	entries := getEntries(statics)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Value == entries[j].Value {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Value > entries[j].Value
	})
	res := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		res = append(res, entries[i].Key)
	}
	return res
}

func getEntries(s map[string]int) []Entry {
	res := make([]Entry, len(s))
	for k, v := range s {
		entry := Entry{
			Key:   k,
			Value: v,
		}
		res = append(res, entry)
	}
	return res
}
