//go:build !solution

package hogwarts

import "slices"

func GetCourseList(prereqs map[string][]string) []string {
	set := make(map[string]struct{}, len(prereqs)*2)
	for c, deps := range prereqs {
		set[c] = struct{}{}
		for _, d := range deps {
			set[d] = struct{}{}
		}
	}

	all := make([]string, 0, len(set))
	for c := range set {
		all = append(all, c)
	}
	slices.Sort(all)

	visited := make(map[string]bool, len(all))
	visiting := make(map[string]bool, len(all))
	order := make([]string, 0, len(all))

	var dfs func(string)
	dfs = func(c string) {
		if visited[c] {
			return
		}
		if visiting[c] {
			panic("cycle detected")
		}
		visiting[c] = true

		for _, d := range prereqs[c] {
			dfs(d)
		}

		visiting[c] = false
		visited[c] = true
		order = append(order, c)
	}

	for _, c := range all {
		dfs(c)
	}
	return order
}