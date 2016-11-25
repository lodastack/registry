package httpd

import (
	"testing"
)

func Test_NewSession(t *testing.T) {
	s := NewSession()
	if s == nil {
		t.Fatalf("new session failed: %v", s)
	}
}

func Test_SetAndGet(t *testing.T) {
	s := NewSession()
	if s == nil {
		t.Fatalf("new session failed: %v", s)
	}
	s.Set("token", "username")
	res := s.Get("token")
	username := res.(string)
	if username != "username" {
		t.Fatalf("get failed failed: %s - %s", username, "username")
	}
}

func Test_Delete(t *testing.T) {
	s := NewSession()
	if s == nil {
		t.Fatalf("new session failed: %v", s)
	}
	s.Set("token", "username")
	s.Delete("token")
	res := s.Get("token")
	if res != nil {
		t.Fatalf("get failed failed: %v", res)
	}
}
