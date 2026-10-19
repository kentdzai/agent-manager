package agentsession

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

type devinRow struct {
	id, cwd           string
	created, activity time.Time
}

func writeDevinStore(t *testing.T, rows ...devinRow) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sessions.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(context.Background(), `CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		working_directory TEXT NOT NULL,
		backend_type TEXT NOT NULL,
		model TEXT NOT NULL,
		agent_mode TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		last_activity_at INTEGER NOT NULL
	)`); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		activity := row.activity
		if activity.IsZero() {
			activity = row.created
		}
		if _, err := db.ExecContext(context.Background(), `INSERT INTO sessions (id, working_directory, backend_type, model, agent_mode, created_at, last_activity_at) VALUES (?, ?, 'cli', 'opus', 'normal', ?, ?)`,
			row.id, row.cwd, row.created.Unix(), activity.Unix()); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

// setDevinActivity stamps a session's activity, as a resumed session's
// first turn does.
func setDevinActivity(t *testing.T, path, id string, at time.Time) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), `UPDATE sessions SET last_activity_at = ? WHERE id = ?`,
		at.Unix(), id); err != nil {
		t.Fatal(err)
	}
}

func TestCaptureDevinPicksSessionAfterLaunchInCwd(t *testing.T) {
	launch := time.Now()
	path := writeDevinStore(t,
		devinRow{id: "old", cwd: "/repo", created: launch.Add(-time.Hour)},
		devinRow{id: "just-before", cwd: "/repo", created: launch.Add(-time.Second)},
		devinRow{id: "other", cwd: "/elsewhere", created: launch.Add(time.Second)},
		devinRow{id: "ours", cwd: "/repo", created: launch.Add(2 * time.Second)},
	)

	id, ok := captureDevin(path, "/repo", launch, map[string]bool{})
	if !ok || id != "ours" {
		t.Fatalf("got id=%q ok=%v, want ours true", id, ok)
	}
}

func TestCaptureDevinSkipsClaimed(t *testing.T) {
	launch := time.Now()
	path := writeDevinStore(t,
		devinRow{id: "first", cwd: "/repo", created: launch.Add(time.Second)},
		devinRow{id: "second", cwd: "/repo", created: launch.Add(2 * time.Second)},
	)

	id, ok := captureDevin(path, "/repo", launch, map[string]bool{"first": true})
	if !ok || id != "second" {
		t.Fatalf("got id=%q ok=%v, want second true", id, ok)
	}
}

func TestCaptureDevinNoMatch(t *testing.T) {
	launch := time.Now()
	path := writeDevinStore(t,
		devinRow{id: "old", cwd: "/repo", created: launch.Add(-time.Hour)},
	)
	if id, ok := captureDevin(path, "/repo", launch, map[string]bool{}); ok {
		t.Fatalf("got id=%q, want no match", id)
	}
	if id, ok := captureDevin(path, "/elsewhere", launch.Add(-time.Hour), map[string]bool{}); ok {
		t.Fatalf("got id=%q, want no match for another cwd", id)
	}
}

func TestRecaptureDevinBindsOnlyWhatOutranTheSnapshot(t *testing.T) {
	launch := time.Now()
	path := writeDevinStore(t,
		devinRow{id: "quiet", cwd: "/repo", created: launch.Add(-time.Hour), activity: launch.Add(-time.Hour)},
		devinRow{id: "picked", cwd: "/repo", created: launch.Add(-time.Hour), activity: launch.Add(-time.Minute)},
		devinRow{id: "other-dir", cwd: "/elsewhere", created: launch.Add(-time.Hour), activity: launch.Add(time.Minute)},
	)
	snapshot, ok := snapshotDevin(path, "/repo")
	if !ok || len(snapshot) != 2 {
		t.Fatalf("snapshot = %v ok=%v, want the two /repo sessions", snapshot, ok)
	}
	setDevinActivity(t, path, "picked", launch.Add(time.Minute))

	cands := recaptureDevin(path, "/repo", snapshot, map[string]bool{})
	if len(cands) != 1 || cands[0].id != "picked" {
		t.Fatalf("cands = %v, want [picked]", cands)
	}
}

func TestRecaptureDevinRefusesTwoThatOutranTheSnapshot(t *testing.T) {
	launch := time.Now()
	path := writeDevinStore(t,
		devinRow{id: "one", cwd: "/repo", created: launch.Add(-time.Hour), activity: launch.Add(-time.Minute)},
		devinRow{id: "two", cwd: "/repo", created: launch.Add(-time.Hour), activity: launch.Add(-time.Minute)},
	)
	snapshot, ok := snapshotDevin(path, "/repo")
	if !ok {
		t.Fatal("snapshot failed")
	}
	setDevinActivity(t, path, "one", launch.Add(time.Minute))
	setDevinActivity(t, path, "two", launch.Add(2*time.Minute))

	cands := recaptureDevin(path, "/repo", snapshot, map[string]bool{})
	if len(cands) != 2 {
		t.Fatalf("cands = %v, want both outran rows so Recapture refuses", cands)
	}
}
