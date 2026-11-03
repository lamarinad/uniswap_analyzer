package metrics

import "testing"

func TestOrderFlowImbalance(t *testing.T) {
	cases := []struct {
		name      string
		buy, sell float64
		want      float64
	}{
		{"покупок втрое больше", 3, 1, 0.5},
		{"поровну", 5, 5, 0},
		{"только продажи", 0, 7, -1},
		{"нет обменов", 0, 0, 0},
	}
	for _, c := range cases {
		if got := OrderFlowImbalance(c.buy, c.sell); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
