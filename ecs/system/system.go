package system

import "github.com/wangtengda0310/gobee/ecs/component"

// 第一层表示system分组，第二层表示system依赖
// 同一组system会在同一update的统一轮迭代中执行
// 不同层级的system会在不同update轮中执行
// 例如 [][]{{1},{3,4}}表示有两组system,第一组system关心componet[1]，第二组system关系component[3], 并在component[3]迭代完成后执行component[4]的迭代
var dispatcher = [][]int{}

func One(p component.Type) {
	var match bool
	for _, round := range dispatcher {
		if len(round) == 0 {
			continue
		}
		if round[0]&int(p) == 0 {
			continue
		} else {
			round[0] |= int(p)
		}
		match = true
	}
	if !match {
		// create new round
		dispatcher = append(dispatcher, []int{int(p)})
	}
}
func Group(p ...component.Type) {
	var mask int
	for _, c := range p {
		mask |= int(c)
	}
	var match bool
	for _, round := range dispatcher {
		if len(round) == 0 {
			continue
		}
		if round[0]&mask == 0 {
			continue
		} else {
			round[0] |= mask
			match = true
		}
	}
	if !match {
		// create new round
		dispatcher = append(dispatcher, []int{mask})
	}

	tryMerge()
}

func tryMerge() {
	for i := 0; i < len(dispatcher); i++ {
		if len(dispatcher[i]) == 0 {
			continue
		}
		for j := i + 1; j < len(dispatcher); j++ {
			if len(dispatcher[j]) == 0 {
				continue
			}
			if dispatcher[i][0]&dispatcher[j][0] != 0 {
				dispatcher[i][0] = dispatcher[i][0] | dispatcher[j][0]
				dispatcher[i] = append(dispatcher[i], dispatcher[j][1:]...)
				dispatcher[j] = dispatcher[len(dispatcher)-1]
				dispatcher = dispatcher[:len(dispatcher)-1]
			}
		}
	}
}

func Round(p ...component.Type) {
	for _, ints := range dispatcher {
		if len(ints) == 0 {
			continue
		}
		for _, i := range ints {
			if int(p[0])&i == 0 {
				continue
			} else {
				ints = append(ints, int(p[0]))
				return
			}
		}
	}
}
