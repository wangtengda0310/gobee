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
	r.items[i].Index = i
	r.items[j].Index = j
}

func (r *Rank) Push(x any) {
	item := x.(*RankItem)
	item.Index = len(r.items)
	r.items = append(r.items, item)
}

func (r *Rank) Pop() any {
	n := len(r.items)
	x := r.items[n-1]
	r.items = r.items[0 : n-1]
	return x
}
