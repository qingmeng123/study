package main

import "fmt"

var (
	path []string
	res  [][]string
	used map[int]bool
)

func findItinerary(tickets [][]string) []string {
	res, path = make([][]string, 0), make([]string, 0)
	used = make(map[int]bool, len(tickets))

	path = append(path, "JFK")
	dfs(tickets)

	return res[0]
}

func dfs(tickets [][]string) {

	if len(path) == len(tickets)+1 {
		tmp := make([]string, len(path))

		copy(tmp, path)
		if len(res) > 0 {
			res[0] = tmp
		}
		res = append(res, tmp)
		return
	}
	for i := 0; i < len(tickets); i++ {

		if used[i] {
			continue
		}
		if len(path) == 1 && tickets[i][0] == "JFK" || path[len(path)-1] == tickets[i][0] {

			if len(res) > 0 && tickets[i][1] > res[0][len(path)] {
				continue
			}
			path = append(path, tickets[i][1])
			used[i] = true
			dfs(tickets)
			fmt.Println(len(path))
			used[i] = false
			path = path[:len(path)-1]
		}
	}
}
