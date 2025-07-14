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
	system    []func(chunk entity.Chunk)
}

var structDispatcher = &systemDispatcher{}

func Group(f func(chunk entity.Chunk), p ...component.Type) {
	var mask int
	for _, c := range p {
		mask |= int(c)
	}
	var match bool
	for _, group := range structDispatcher.group {
		if len(group.round) == 0 {
			continue
		}
		if group.round[0].archetype != mask {
			continue
		} else {
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
					system:    []func(chunk entity.Chunk){f},
				},
			},
		})
	}

}

func Round(p component.Type, q component.Type, f func(chunk entity.Chunk)) {
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
					system:    []func(chunk entity.Chunk){f},
				})
			}
			return
		}
	}
	structDispatcher.group = append(structDispatcher.group, DispatcherGroup{
		round: []DispatcherRound{
			{
				archetype: int(p),
				system:    []func(chunk entity.Chunk){f},
			},
			{
				archetype: int(q),
				system:    []func(chunk entity.Chunk){f},
			}}})
}

func Range(dispatcher systemDispatcher) int {
	var c int
	for archetype, chunk := range entity.Chunks {
		for _, groupedSystem := range dispatcher.group {
			for _, sc := range groupedSystem.round {
				if int(archetype)&sc.archetype == sc.archetype {
					c++

					// not grouped can async
					var args []any
					func(c int, chunk entity.Chunk) {
						// grouped must sync
						for _, f := range sc.system {
							f(chunk)
						}
						println("call system with args", args)
						//println("Entity:", i, "Dispatcher:", &sc, "Group", c, "type:", chunk.Archetype)
					}(c, *chunk)
				}
			}
		}

	}
	return c
}

func Update() int {
	return Range(*structDispatcher)
}
