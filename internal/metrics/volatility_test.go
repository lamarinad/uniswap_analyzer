package metrics

import (
	"math"
	"testing"
)

func TestStepLogReturns(t *testing.T) {
	// 100 -> 110 -> 99: два шага
	got := StepLogReturns([]float64{100, 110, 99})
	want := []float64{math.Log(1.1), math.Log(0.9)}
	if len(got) != len(want) {
		t.Fatalf("len got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Fatalf("got[%d]=%v, want %v", i, got[i], want[i])
		}
	}
	if got := StepLogReturns([]float64{100}); got != nil {
		t.Fatalf("one price: got %v, want nil", got)
	}
}

func TestRealizedVol(t *testing.T) {
	// sqrt(0.03^2 + 0.04^2) = 0.05
	got := RealizedVol([]float64{0.03, -0.04})
	if math.Abs(got-0.05) > 1e-12 {
		t.Fatalf("got %v, want 0.05", got)
	}
}

func TestRealizedVolCalmVsShaky(t *testing.T) {
	// оба пути приходят к одной цене 101, но второй мечется - его волатильность выше
	calm := RealizedVol(StepLogReturns([]float64{100, 100.5, 101}))
	shaky := RealizedVol(StepLogReturns([]float64{100, 103, 98, 101}))
	if shaky <= calm {
		t.Fatalf("shaky %v should be > calm %v", shaky, calm)
	}
}
