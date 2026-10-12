package main

import "testing"

func (r repo) wantMake(target, want string) {
	r.t.Helper()
	if got := r.make(target); got != want {
		r.t.Errorf("make %s = %q, ожидалось %q", target, got, want)
	}
}

func TestMakeCallsGoReleaser(t *testing.T) {
	t.Parallel()
	r := newRepo(t)
	r.wantMake("build", "build --clean --single-target --auto-snapshot --snapshot -o stonemcp")
	r.wantMake("release", "build --clean --auto-snapshot --snapshot")
	r.tag("27.0")
	r.wantMake("build", "build --clean --single-target --auto-snapshot -o stonemcp")
	r.wantMake("release", "build --clean --auto-snapshot")
	r.untracked()
	r.wantMake("build", "build --clean --single-target --auto-snapshot -o stonemcp")
}
