package likes

import "testing"

func TestClampPaging(t *testing.T) {
	if page, size := clampPaging(0, MaxPageSize+1); page != 1 || size != MaxPageSize {
		t.Fatalf("clampPaging = (%d, %d)", page, size)
	}
}
