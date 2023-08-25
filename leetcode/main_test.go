package main

import (
	"fmt"
	"testing"
)

func Test_combine(t *testing.T) {
	ti := make([][]string, 0)
	ti = [][]string{{"EZE", "AXA"}, {"TIA", "ANU"}, {"ANU", "JFK"}, {"JFK", "ANU"}, {"ANU", "EZE"}, {"TIA", "ANU"}, {"AXA", "TIA"}, {"TIA", "JFK"}, {"ANU", "TIA"}, {"JFK", "TIA"}}
	ti = [][]string{{"JFK", "SFO"}, {"JFK", "ATL"}, {"SFO", "ATL"}, {"ATL", "JFK"}, {"ATL", "SFO"}}
	fmt.Println(findItinerary(ti))
}

var m = [8]string{"abc", "def", "ghi", "jkl", "mno", "pqrs", "tuv", "wxyz"}
