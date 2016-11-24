package dup

import (
	"reflect"
	"strings"
	"testing"
)

func Test_suppressedDates(t *testing.T) {
	tests := []struct {
		notes []string
		want  []string
	}{
		{
			notes: []string{},
			want:  nil,
		},
		{
			notes: []string{"SuppressDuplicates: 2016/04/03"},
			want:  []string{"2016/04/03"},
		},
		{
			notes: []string{"SuppressDuplicates: 2016/04/03, 2016/04/04,2016/04/05"},
			want:  []string{"2016/04/03", "2016/04/04", "2016/04/05"},
		},
		{
			notes: []string{
				"SuppressDuplicates: 2016/02/03",
				"some irrelevant stuff here: 2016/02/04",
				"SuppressDuplicates: 2016/02/05"},
			want: []string{"2016/02/03", "2016/02/05"},
		},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.notes, "_"), func(t *testing.T) {
			if got := suppressedDates(tt.notes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("suppressedDates() = %v, want %v", got, tt.want)
			}
		})
	}
}
