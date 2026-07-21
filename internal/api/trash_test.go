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
