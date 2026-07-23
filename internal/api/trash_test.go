package api

import (
	"testing"

	"github.com/tanq16/box/internal/types"
)

func TestSelectTrashItem(t *testing.T) {
	items := []types.BoxItemDisplay{
		{ID: "1", Name: "report.pdf", Type: "file"},
		{ID: "2", Name: "Photos", Type: "folder"},
		{ID: "3", Name: "photos", Type: "folder"},
	}

	tests := []struct {
		name     string
		argName  string
		argID    string
		wantID   string
		wantType string
		wantErr  bool
	}{
		{name: "by id", argID: "2", wantID: "2", wantType: "folder"},
		{name: "id not in trash", argID: "99", wantErr: true},
		{name: "unique name", argName: "report.pdf", wantID: "1", wantType: "file"},
		{name: "name is case insensitive but ambiguous", argName: "PHOTOS", wantErr: true},
		{name: "no match", argName: "missing", wantErr: true},
		{name: "id wins over name", argName: "photos", argID: "1", wantID: "1", wantType: "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, itemType, err := selectTrashItem(items, tt.argName, tt.argID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got id=%q type=%q", id, itemType)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID || itemType != tt.wantType {
				t.Fatalf("got id=%q type=%q, want id=%q type=%q", id, itemType, tt.wantID, tt.wantType)
			}
		})
	}
}

func countAction(ops []SyncOp, a SyncAction) int {
	n := 0
	for _, op := range ops {
		if op.Action == a {
			n++
		}
	}
	return n
}

func TestCollectSyncOps(t *testing.T) {
	local := &types.FileTree{
		Files: map[string]types.FileInfo{
			"a.txt":    {Hash: "1"},
			"b.txt":    {Hash: "2"},
			"same.txt": {Hash: "9"},
			"gone.txt": {Hash: "x", ID: "fg"},
		},
		Dirs: map[string]*types.FileTree{
			"newdir": {Files: map[string]types.FileInfo{"c.txt": {Hash: "3"}}, Dirs: map[string]*types.FileTree{}},
			"olddir": {Files: map[string]types.FileInfo{}, Dirs: map[string]*types.FileTree{}},
		},
	}
	remote := &types.FileTree{
		Files: map[string]types.FileInfo{
			"b.txt":    {Hash: "OLD", ID: "fb"},
			"same.txt": {Hash: "9"},
			"del.txt":  {Hash: "z", ID: "fd"},
		},
		Dirs: map[string]*types.FileTree{
			"olddir":     {ID: "do", Files: map[string]types.FileInfo{}, Dirs: map[string]*types.FileTree{}},
			"remoteonly": {ID: "ro", Files: map[string]types.FileInfo{"r.txt": {Hash: "7", ID: "fr"}}, Dirs: map[string]*types.FileTree{}},
		},
	}

	push := collectPushOps(local, remote, "")
	if got := countAction(push, SyncCreateFolder); got != 1 {
		t.Fatalf("push create-folder count = %d, want 1", got)
	}
	if got := countAction(push, SyncDeleteFolder); got != 1 {
		t.Fatalf("push delete-folder count = %d, want 1", got)
	}
	pAdd, pUpdate, pDel, pFolders := summarize(push)
	if pAdd != 3 || pUpdate != 1 || pDel != 1 || pFolders != 2 {
		t.Fatalf("push summary = add %d update %d del %d folders %d, want 3 1 1 2", pAdd, pUpdate, pDel, pFolders)
	}

	pull := collectPullOps(remote, local, "")
	if got := countAction(pull, SyncCreateFolder); got != 0 {
		t.Fatalf("pull must not emit create-folder ops, got %d", got)
	}
	uAdd, uUpdate, uDel, uFolders := summarize(pull)
	if uAdd != 2 || uUpdate != 1 || uDel != 2 || uFolders != 1 {
		t.Fatalf("pull summary = add %d update %d del %d folders %d, want 2 1 2 1", uAdd, uUpdate, uDel, uFolders)
	}
}

func TestSyncPlanDeletes(t *testing.T) {
	clean := &SyncPlan{Ops: []SyncOp{{SyncAdd, "a"}, {SyncUpdate, "b"}}}
	if clean.HasDeletes() || clean.DeleteTotal() != 0 {
		t.Fatalf("clean plan reported deletes: total=%d", clean.DeleteTotal())
	}
	dirty := &SyncPlan{Ops: []SyncOp{{SyncAdd, "a"}, {SyncDelete, "b"}, {SyncDeleteFolder, "c"}}}
	if !dirty.HasDeletes() || dirty.DeleteTotal() != 2 {
		t.Fatalf("dirty plan delete total = %d, want 2", dirty.DeleteTotal())
	}
}
