package tags

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	tg := New()
	tg.Set("env", "prod")
	v, ok := tg.Get("env")
	if !ok || v != "prod" {
		t.Fatalf("expected prod, got %q ok=%v", v, ok)
	}
}

func TestGet_Missing(t *testing.T) {
	tg := New()
	_, ok := tg.Get("missing")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestDelete(t *testing.T) {
	tg := New()
	tg.Set("k", "v")
	tg.Delete("k")
	_, ok := tg.Get("k")
	if ok {
		t.Fatal("expected key deleted")
	}
}

func TestKeys_Sorted(t *testing.T) {
	tg := New()
	tg.Set("z", "1")
	tg.Set("a", "2")
	tg.Set("m", "3")
	keys := tg.Keys()
	if keys[0] != "a" || keys[1] != "m" || keys[2] != "z" {
		t.Fatalf("unexpected order: %v", keys)
	}
}

func TestClone_IsIndependent(t *testing.T) {
	tg := New()
	tg.Set("x", "1")
	cloned := tg.Clone()
	cloned.Set("x", "99")
	v, _ := tg.Get("x")
	if v != "1" {
		t.Fatal("original should not be mutated")
	}
}

func TestMatches_AllPresent(t *testing.T) {
	tg := New()
	tg.Set("env", "prod")
	tg.Set("region", "us-east")
	filter := New()
	filter.Set("env", "prod")
	if !tg.Matches(filter) {
		t.Fatal("expected match")
	}
}

func TestMatches_Missing(t *testing.T) {
	tg := New()
	tg.Set("env", "prod")
	filter := New()
	filter.Set("env", "staging")
	if tg.Matches(filter) {
		t.Fatal("expected no match")
	}
}

func TestMatches_EmptyFilter(t *testing.T) {
	tg := New()
	tg.Set("env", "prod")
	if !tg.Matches(New()) {
		t.Fatal("empty filter should match everything")
	}
}
