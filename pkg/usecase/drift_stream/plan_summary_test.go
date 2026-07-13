package drift_stream

import "testing"

func TestParsePlanSummary(t *testing.T) {
	tests := []struct {
		name                      string
		output                    string
		wantMatch                 bool
		wantAdd, wantChg, wantDst int32
	}{
		{
			name:      "drift with single change",
			output:    "OpenTofu will perform the following actions:\n\nPlan: 0 to add, 1 to change, 0 to destroy.\n",
			wantMatch: true,
			wantAdd:   0, wantChg: 1, wantDst: 0,
		},
		{
			name:      "all counts populated",
			output:    "Plan: 3 to add, 2 to change, 1 to destroy.",
			wantMatch: true,
			wantAdd:   3, wantChg: 2, wantDst: 1,
		},
		{
			name:      "no changes",
			output:    "No changes. Your infrastructure matches the configuration.",
			wantMatch: false,
		},
		{
			name:      "empty output",
			output:    "",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, changed, destroyed := ParsePlanSummary(tt.output)
			if !tt.wantMatch {
				if added != nil || changed != nil || destroyed != nil {
					t.Fatalf("expected nil counts, got %v/%v/%v", added, changed, destroyed)
				}
				return
			}
			if added == nil || changed == nil || destroyed == nil {
				t.Fatalf("expected non-nil counts, got %v/%v/%v", added, changed, destroyed)
			}
			if *added != tt.wantAdd || *changed != tt.wantChg || *destroyed != tt.wantDst {
				t.Errorf("got %d/%d/%d, want %d/%d/%d", *added, *changed, *destroyed, tt.wantAdd, tt.wantChg, tt.wantDst)
			}
		})
	}
}
