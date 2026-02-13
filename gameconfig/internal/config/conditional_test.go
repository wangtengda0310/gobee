package config

import (
	"testing"
)

// TestMapRowWithConditions 测试带条件的行映射
func TestMapRowWithConditions(t *testing.T) {
	type TestItem struct {
		Type    int    `excel:"type"`
		Name     string `excel:"name,required"`
		Attack   int    `excel:"attack,when:type=1"`
		Defense  int    `excel:"defense,when:type=2"`
		Speed    int    `excel:"speed,when:type>0"`
		Hidden   string `excel:"hidden,when:type in [1,3]"`
	}

	tests := []struct {
		name   string
		headers []string
		row    []string
		want   TestItem
	}{
		{
			name:   "type=0, 普通道具",
			headers: []string{"type", "name", "attack", "defense", "speed", "hidden"},
			row:    []string{"0", "普通物品", "100", "50", "5", "secret"},
			want:   TestItem{Type: 0, Name: "普通物品", Attack: 0, Defense: 0, Speed: 0, Hidden: ""},
		},
		{
			name:   "type=1, 武器",
			headers: []string{"type", "name", "attack", "defense", "speed", "hidden"},
			row:    []string{"1", "铁剑", "100", "50", "5", "secret"},
			want:   TestItem{Type: 1, Name: "铁剑", Attack: 100, Defense: 0, Speed: 5, Hidden: "secret"},
		},
		{
			name:   "type=2, 盔甲",
			headers: []string{"type", "name", "attack", "defense", "speed", "hidden"},
			row:    []string{"2", "铁甲", "100", "50", "5", "secret"},
			want:   TestItem{Type: 2, Name: "铁甲", Attack: 0, Defense: 50, Speed: 5, Hidden: ""}, // Hidden 条件不满足 (type in [1,3])
		},
		{
			name:   "type=3, 特殊武器",
			headers: []string{"type", "name", "attack", "defense", "speed", "hidden"},
			row:    []string{"3", "特殊剑", "100", "50", "5", "secret"},
			want:   TestItem{Type: 3, Name: "特殊剑", Attack: 0, Defense: 0, Speed: 5, Hidden: "secret"}, // Attack 条件不满足 (type=1)，Hidden 满足 (type in [1,3])
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewStructMapper[TestItem]()

			result, err := mapper.MapRow(tt.headers, tt.row)
			if err != nil {
				t.Fatalf("MapRow 失败: %v", err)
			}

			if result.Type != tt.want.Type {
				t.Errorf("Type = %d, want %d", result.Type, tt.want.Type)
			}
			if result.Name != tt.want.Name {
				t.Errorf("Name = %s, want %s", result.Name, tt.want.Name)
			}
			if result.Attack != tt.want.Attack {
				t.Errorf("Attack = %d, want %d", result.Attack, tt.want.Attack)
			}
			if result.Defense != tt.want.Defense {
				t.Errorf("Defense = %d, want %d", result.Defense, tt.want.Defense)
			}
			if result.Speed != tt.want.Speed {
				t.Errorf("Speed = %d, want %d", result.Speed, tt.want.Speed)
			}
			if result.Hidden != tt.want.Hidden {
				t.Errorf("Hidden = %s, want %s", result.Hidden, tt.want.Hidden)
			}
		})
	}
}
