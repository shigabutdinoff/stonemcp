package fakegame

import (
	"fmt"
	"net"
	"slices"
	"sync"
	"testing"
	"time"
)

const sep = "\x1f"

type recorder struct {
	testing.TB
	mu   sync.Mutex
	errs []string
}

func (r *recorder) Helper() {}

func (r *recorder) Error(args ...any) { r.add(fmt.Sprint(args...)) }

func (r *recorder) Errorf(format string, args ...any) { r.add(fmt.Sprintf(format, args...)) }

func (r *recorder) add(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errs = append(r.errs, msg)
}

func (r *recorder) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.errs)
}

func recv(t *testing.T, conn net.Conn) string {
	t.Helper()
	buf := make([]byte, 4096)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}

func exchange(t *testing.T, conn net.Conn, packet string) string {
	t.Helper()
	recv(t, conn)
	if _, err := conn.Write([]byte(packet)); err != nil {
		t.Fatal(err)
	}
	return recv(t, conn)
}
