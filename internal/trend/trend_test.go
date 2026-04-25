package trend

import (
	"testing"
	"time"
)

func TestDirection_FlatWhenNoSamples(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 3})
	if got := tr.Direction("k"); got != Flat {
		t.Fatalf("expected Flat, got %s", got)
	}
}

func TestDirection_FlatBelowMinSamples(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 4})
	for i := 0; i < 3; i++ {
		tr.Record("k", 1.0)
	}
	if got := tr.Direction("k"); got != Flat {
		t.Fatalf("expected Flat below min samples, got %s", got)
	}
}

func TestDirection_Rising(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 4})
	for _, v := range []float64{1, 2, 3, 4} {
		tr.Record("k", v)
	}
	if got := tr.Direction("k"); got != Rising {
		t.Fatalf("expected Rising, got %s", got)
	}
}

func TestDirection_Falling(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 4})
	for _, v := range []float64{4, 3, 2, 1} {
		tr.Record("k", v)
	}
	if got := tr.Direction("k"); got != Falling {
		t.Fatalf("expected Falling, got %s", got)
	}
}

func TestDirection_IndependentKeys(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 2})
	tr.Record("up", 1)
	tr.Record("up", 2)
	tr.Record("down", 2)
	tr.Record("down", 1)
	if got := tr.Direction("up"); got != Rising {
		t.Errorf("up: expected Rising, got %s", got)
	}
	if got := tr.Direction("down"); got != Falling {
		t.Errorf("down: expected Falling, got %s", got)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	tr := New(Config{Window: time.Minute, MinSamples: 2})
	tr.Record("k", 1)
	tr.Record("k", 2)
	tr.Reset("k")
	if got := tr.Direction("k"); got != Flat {
		t.Fatalf("expected Flat after reset, got %s", got)
	}
}

func TestDirectionString(t *testing.T) {
	cases := map[Direction]string{
		Flat:    "flat",
		Rising:  "rising",
		Falling: "falling",
	}
	for d, want := range cases {
		if got := d.String(); got != want {
			t.Errorf("Direction(%d).String() = %q, want %q", d, got, want)
		}
	}
}

func TestDefaults_Applied(t *testing.T) {
	tr := New(Config{})
	if tr.cfg.Window != time.Minute {
		t.Errorf("expected default window 1m, got %s", tr.cfg.Window)
	}
	if tr.cfg.MinSamples != 3 {
		t.Errorf("expected default MinSamples 3, got %d", tr.cfg.MinSamples)
	}
}
