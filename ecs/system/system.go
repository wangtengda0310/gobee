package system

import (
	"github.com/wangtengda0310/gobee/ecs/component"
	"github.com/wangtengda0310/gobee/ecs/entity"
)

// 第一层表示system分组，第二层表示system依赖
// 同一组system会在同一update的统一轮迭代中执行
// 不同层级的system会在不同update轮中执行
// 例如 [][]{{1},{3,4}}表示有两组system,第一组system关心componet[1]，第二组system关系component[3], 并在component[3]迭代完成后执行component[4]的迭代

type systemDispatcher struct {
	group []DispatcherGroup
}

type DispatcherGroup struct {
	round []DispatcherRound
}
type DispatcherRound struct {
	archetype int
	system    []func()
}

var structDispatcher = &systemDispatcher{}

func One(p component.Type, system func()) {
	var match bool
	for _, group := range structDispatcher.group {
		if len(group.round) == 0 {
			continue
		}
		if group.round[0].archetype&int(p) == 0 {
			continue
		} else {
			group.round[0].archetype |= int(p)
			match = true
		}
	}
	if !match {
		// create new round
		structDispatcher.group = append(structDispatcher.group, DispatcherGroup{
			round: []DispatcherRound{
				{
					archetype: int(p),
					system:    []func(){system},
				},
			},
		})
	}
}

func Group(f func(), p ...component.Type) {
	var mask int
	for _, c := range p {
		mask |= int(c)
	}
	var match bool
	for _, group := range structDispatcher.group {
		if len(group.round) == 0 {
			continue
		}
		if group.round[0].archetype&mask == 0 {
			continue
		} else {
			group.round[0].archetype |= mask
			group.round[0].system = append(group.round[0].system, f)
			match = true
		}
	}
	if !match {
		// create new round
		structDispatcher.group = append(structDispatcher.group, DispatcherGroup{
			round: []DispatcherRound{
				{
					archetype: mask,
					system:    []func(){f},
				},
			},
		})
	}

	tryMerge()
}

func tryMerge() {
	for g, currentGroup := range structDispatcher.group {
		if len(currentGroup.round) == 0 {
			continue
		}
		if g == len(structDispatcher.group)-1 {
			// last group, no next group to merge
			continue
		}
		nextGroup := structDispatcher.group[g+1]
		nextGroupAllRound := nextGroup.round
		if nextGroupAllRound[0].archetype&currentGroup.round[0].archetype != 0 {
			var r = 0
			// todo 这里使用树状结构是否可以使round易于并行执行
			for r < len(nextGroup.round) {
				currentGroup.round[r].archetype |= nextGroupAllRound[r].archetype
				currentGroup.round[r].system = append(currentGroup.round[r].system, nextGroup.round[r].system...)
				r++
			}
		}
	}
}

func Round(p component.Type, q component.Type, f func()) {
	for _, group := range structDispatcher.group {
		if len(group.round) == 0 {
			continue
		}
		if group.round[0].archetype&int(p) == 0 {
			continue
		} else {
			group.round[0].archetype |= int(p)
			if len(group.round) > 1 {
				group.round[1].system = append(group.round[1].system, f)
			} else {
				group.round = append(group.round, DispatcherRound{
					archetype: int(q),
					system:    []func(){f},
				})
			}
			return
		}
	}
	structDispatcher.group = append(structDispatcher.group, DispatcherGroup{
		round: []DispatcherRound{
			{
				archetype: int(p),
				system:    []func(){f},
			},
			{
				archetype: int(q),
				system:    []func(){f},
			}}})
}

func Range(dispatcher systemDispatcher) int {
	var c int
	for i := 1; i < entity.L; i++ {
		ec := entity.Pool[i]
		for _, components := range dispatcher.group {
			for _, sc := range components.round {
				if ec&sc.archetype == sc.archetype {
					c++
					// load components
					var eca [][]component.Component
					var cce = 1
					// grouped must sync
					for cce <= sc.archetype {
						data := component.Data(component.Type(cce))
						if cce > len(eca)-1 {
							// expand eca
							newEca := make([][]component.Component, cce+1)
							copy(newEca, eca)
							eca = newEca
						}
						eca[cce] = data
						cce = cce << 1
					}

					// not grouped can async
					var args []any
					go func(c int) {
						var cc = 1
						// grouped must sync
						for cc <= sc.archetype {
							if cc&sc.archetype == 0 {
								continue
							}
							args = append(args, eca[cc][0])
							for _, f := range sc.system {
								f()
							}
							println("call system with args", args)
							println("Entity:", i, "Components:", ec, "Dispatcher:", &sc, "Group", c, "type:", cc)
							cc = cc << 1
						}
					}(c)
				}
			}
		}
	}
	return c
}
