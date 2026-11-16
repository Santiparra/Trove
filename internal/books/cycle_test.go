package books

import "testing"

func TestLiveReadingDates(t *testing.T) {
	for _, test := range []struct {
		name                  string
		before                Cycle
		next                  string
		editStart, editFinish bool
		wantStart, wantFinish string
		fails                 bool
	}{
		{name: "start", before: Cycle{Status: "to-read"}, next: "reading", wantStart: "2026-06-10"},
		{name: "resume", before: Cycle{Status: "paused", StartedOn: "2026-01-01"}, next: "reading", wantStart: "2026-01-01"},
		{name: "finish reading", before: Cycle{Status: "reading", StartedOn: "2026-01-01"}, next: "read", wantStart: "2026-01-01", wantFinish: "2026-06-10"},
		{name: "finish paused", before: Cycle{Status: "paused", StartedOn: "2026-01-01"}, next: "read", wantStart: "2026-01-01", wantFinish: "2026-06-10"},
		{name: "finish abandoned", before: Cycle{Status: "did-not-finish", StartedOn: "2026-01-01"}, next: "read", wantStart: "2026-01-01", wantFinish: "2026-06-10"},
		{name: "finish pending with known start", before: Cycle{Status: "to-read", StartedOn: "2026-01-01"}, next: "read", wantStart: "2026-01-01", wantFinish: "2026-06-10"},
		{name: "historical read", before: Cycle{Status: "to-read"}, next: "read"},
		{name: "already read", before: Cycle{Status: "read", StartedOn: "2026-01-01"}, next: "read", wantStart: "2026-01-01"},
		{name: "known completion", before: Cycle{Status: "paused", StartedOn: "2026-01-01", FinishedOn: "2026-02-01"}, next: "read", wantStart: "2026-01-01", wantFinish: "2026-02-01"},
		{name: "same day", before: Cycle{Status: "reading", StartedOn: "2026-06-10"}, next: "read", wantStart: "2026-06-10", wantFinish: "2026-06-10"},
		{name: "future start", before: Cycle{Status: "reading", StartedOn: "2026-07-01"}, next: "read", fails: true},
		{name: "manual clear start", before: Cycle{Status: "to-read"}, next: "reading", editStart: true},
		{name: "manual clear finish", before: Cycle{Status: "reading", StartedOn: "2026-01-01"}, next: "read", editFinish: true, wantStart: "2026-01-01"},
		{name: "manual up next", before: Cycle{Status: "to-read"}, next: "up-next"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.before.Transition(test.next, "2026-06-10", test.editStart, test.editFinish)
			if (err != nil) != test.fails {
				t.Fatalf("error inesperado: %v", err)
			}
			if !test.fails && (got.Status != test.next || got.StartedOn != test.wantStart || got.FinishedOn != test.wantFinish) {
				t.Fatalf("fechas incorrectas: %+v", got)
			}
		})
	}
}
