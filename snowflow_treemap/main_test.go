package main

import (
	"math"
	"testing"
)

func Test_args_transform(t *testing.T) {
	type fields struct {
		segmentsName []string
		segmentsbit  []uint8
	}
	type args struct {
		id uint32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "one segment",
			fields: fields{
				segmentsName: []string{"whole id"},
				segmentsbit:  []uint8{32},
			},
			args: args{
				id: math.MaxInt32,
			},
			want: "2147483647",
		},
		{
			"simple",
			fields{
				segmentsName: []string{"a", "b", "c"},
				segmentsbit:  []uint8{1, 1, 30},
			},
			args{7},
			"0/0/7",
		},
		{
			name: "text max uint32",
			fields: fields{
				segmentsName: []string{"a", "b"},
				segmentsbit:  []uint8{1, 31},
			},
			args: args{math.MaxUint32},
			want: "1/2147483647",
		},
		{
			name: "segments bit lack",
			fields: fields{
				segmentsName: []string{"a"},
				segmentsbit:  []uint8{1},
			},
			args: args{
				id: math.MaxUint32,
			},
			want: "1/2147483647",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ParsingArgs{
				segmentsName: tt.fields.segmentsName,
				segmentsBits: tt.fields.segmentsbit,
			}
			if got := a.transform(tt.args.id); got != tt.want {
				t.Errorf("transform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mask(t *testing.T) {
	type args struct {
		bits   uint8
		offset int
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "highest 1 bit",
			args: args{
				bits:   1,
				offset: 31,
			},
			want: 0b10000000_00000000_00000000_00000000,
		},
		{
			name: "low 31 bit",
			args: args{
				bits:   31,
				offset: 0,
			},
			want: 0b011111111_11111111_11111111_1111111,
		},
		{
			name: "highest 8 bit",
			args: args{
				bits:   8,
				offset: 24,
			},
			want: 0b11111111_00000000_00000000_00000000,
		},
		{
			name: "whole 32 bit",
			args: args{
				bits:   32,
				offset: 0,
			},
			want: 0b11111111_11111111_11111111_11111111,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mask(tt.args.bits, tt.args.offset); got != tt.want {
				t.Errorf("mask() = %v, want %v", got, tt.want)
			}
		})
	}
}
