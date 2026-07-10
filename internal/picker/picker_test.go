package picker

import "testing"

func TestResolvePicker(t *testing.T) {
	tests := []struct {
		name      string
		mode      Mode
		available bool
		wantFzf   bool
		wantErr   bool
	}{
		{"auto with fzf", Auto, true, true, false},
		{"auto without fzf", Auto, false, false, false},
		{"builtin ignores fzf", Builtin, true, false, false},
		{"fzf present", Fzf, true, true, false},
		{"fzf absent errors", Fzf, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePicker(tt.mode, tt.available)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantFzf {
				t.Fatalf("useFzf = %v, want %v", got, tt.wantFzf)
			}
		})
	}
}
