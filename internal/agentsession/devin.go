package agentsession

import (
	"context"
	"path/filepath"
	"time"
)

// devinStateDB is the sessions store Devin keeps under its per-user data
// directory, <data dir>/devin/cli/sessions.db. Its sessions table records
// the working directory and unix-second timestamps, which is all a capture
// needs: created_at names the launch and last_activity_at the turn that
// followed a resume.
func devinStateDB() string {
	root := devinDataDir()
	if root == "" {
		return ""
	}
	return filepath.Join(root, "cli", "sessions.db")
}

func captureDevin(path, cwd string, launchedAt time.Time, claimed map[string]bool) (string, bool) {
	return captureCandidates(devinCandidates(path, cwd, launchedAt, claimed))
}

func snapshotDevin(path, cwd string) (map[string]int64, bool) {
	return snapshotCandidates(devinCandidates(path, cwd, time.Time{}, nil))
}

func recaptureDevin(path, cwd string, snapshot map[string]int64, claimed map[string]bool) []candidate {
	cands, err := devinCandidates(path, cwd, time.Time{}, claimed)
	return recaptureCandidates(cands, err, snapshot)
}

// devinCandidates reads the sessions Devin recorded in cwd. cutoff bounds
// capture to sessions minted at or after the launch; the zero time admits
// every session, which is what snapshot and recapture read against. Both
// columns are whole seconds, so the cutoff compares at second granularity:
// devin's own row, written in the launch's second, is never floor-truncated
// out.
func devinCandidates(path, cwd string, cutoff time.Time, claimed map[string]bool) ([]candidate, error) {
	db, err := openReadOnly(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, `
		SELECT id, working_directory, created_at, last_activity_at
		FROM sessions`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	wantCwd := resolvePath(cwd)
	var cands []candidate
	for rows.Next() {
		var id, sessionCwd string
		var created, activity int64
		if err := rows.Scan(&id, &sessionCwd, &created, &activity); err != nil {
			return nil, err
		}
		if !sessionIDPattern.MatchString(id) || resolvePath(sessionCwd) != wantCwd {
			continue
		}
		modTime := time.Unix(activity, 0)
		if !cutoff.IsZero() {
			if claimed[id] || created < cutoff.Unix() {
				continue
			}
			modTime = time.Unix(created, 0)
		}
		cands = append(cands, candidate{id: id, modTime: modTime})
	}
	return cands, rows.Err()
}
