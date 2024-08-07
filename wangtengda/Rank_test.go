package main

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	rank := Rank{}
	rank.Push(&RankItem{1, "a", 1})
	rank.Push(&RankItem{1, "b", 2})
	rank.Push(&RankItem{1, "d", 4})
	rank.Push(&RankItem{1, "c", 3})
	assert.Equal(t, 4, rank.Len())
	heap.Init(&rank)
	assert.Equal(t, []*RankItem{
		{1, "a", 1},
		{1, "b", 2},
		{1, "c", 3},
		{1, "d", 4},
	}, rank.items)
	assert.Equal(t, "a", rank.Pop().(*RankItem).Name)
	assert.Equal(t, "b", rank.Pop().(*RankItem).Name)
	assert.Equal(t, "c", rank.Pop().(*RankItem).Name)
	assert.Equal(t, "d", rank.Pop().(*RankItem).Name)
}
