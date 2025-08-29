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
	words := strings.Fields(str)
	if len(words) == 0 {
		return make([]string, 0)
	}
	statics := collectStatistics(words)
	entries := getEntries(statics)
	sortEntriesByWordCountAndLexicographically(entries)
	return collectTop10Slice(entries)
}

func sortEntriesByWordCountAndLexicographically(entries []Entry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Value == entries[j].Value {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Value > entries[j].Value
	})
}

func collectStatistics(words []string) map[string]int {
	statics := make(map[string]int, len(words))
	for _, v := range words {
		statics[v]++
	}
	return statics
}

func collectTop10Slice(entries []Entry) []string {
	res := make([]string, 0, 10)
	var l int
	if len(entries) < 10 {
		l = len(entries)
	} else {
		l = 10
	}
	for i := 0; i < l; i++ {
		if entries[i].Key != "" {
			res = append(res, entries[i].Key)
		}
	}
	return res
}

func getEntries(s map[string]int) []Entry {
	res := make([]Entry, 0, len(s))
	for k, v := range s {
		entry := Entry{
			Key:   k,
			Value: v,
		}
		res = append(res, entry)
	}
	return res
}
