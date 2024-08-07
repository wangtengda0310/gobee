package main

type Rank struct {
	items []*RankItem
}
type RankItem struct {
	Index int
	Name  string
	Score int
}

func (r *Rank) Len() int {
	return len(r.items)
}

func (r *Rank) Less(i, j int) bool {
	return r.items[i].Score < r.items[j].Score
}

func (r *Rank) Swap(i, j int) {
	r.items[i], r.items[j] = r.items[j], r.items[i]
}

func (r *Rank) Push(x any) {
	r.items = append(r.items, x.(*RankItem))
}

func (r *Rank) Pop() any {
	n := len(r.items)
	x := r.items[n-1]
	r.items = r.items[:len(r.items)-1]
	return x
}
