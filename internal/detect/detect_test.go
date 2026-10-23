package detect

import "testing"

func TestAvailable(t *testing.T) {
	if !Available("sh") {
		t.Fatal("expected sh to be available on the test host")
	}
	if Available("firstboot-command-that-does-not-exist") {
		t.Fatal("unexpected command reported as available")
	}
}
