package metrics

import (
	"math"
	"testing"
)

func TestSimpleReturn(t *testing.T) {
	// 3200 -> 3232 = +1%
	got := SimpleReturn(3200, 3232)
	if math.Abs(got-0.01) > 1e-12 {
		t.Fatalf("got %v, want 0.01", got)
	}
	if got := SimpleReturn(0, 3232); got != 0 {
		t.Fatalf("zero start: got %v, want 0", got)
	}
}

func TestLogReturnAdds(t *testing.T) {
	// 100 -> 110 -> 99: сумма двух шагов равна доходности за весь путь
	sum := LogReturn(100, 110) + LogReturn(110, 99)
	whole := LogReturn(100, 99)
	if math.Abs(sum-whole) > 1e-12 {
		t.Fatalf("sum %v != whole %v", sum, whole)
	}
}
