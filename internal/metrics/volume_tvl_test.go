package metrics

import "testing"

func TestVolumeTVL(t *testing.T) {
	// за час наторговали 4 млн при 80 млн в пуле -> 0.05 (5% денег пула)
	got, ok := VolumeTVL(4_000_000, 80_000_000)
	if !ok || got != 0.05 {
		t.Fatalf("got %v, %v, want 0.05, true", got, ok)
	}

	// пул пустой значит делить не на что
	if _, ok := VolumeTVL(1000, 0); ok {
		t.Fatal("TVL = 0: want ok = false")
	}
}
