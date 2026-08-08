package port

import "testing"

func TestValidate(t *testing.T) {
	for _, s := range []string{"1", "80", "65535"} {
		if err := Validate(s); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", s, err)
		}
	}
	for _, s := range []string{"0", "-1", "65536", "abc", "", "12.5"} {
		if err := Validate(s); err == nil {
			t.Errorf("Validate(%q) = nil, want error", s)
		}
	}
}
