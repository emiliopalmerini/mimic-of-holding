package vault

import (
	"testing"
)

// --- Unit tests: parsing individual folder names ---

func TestParseScopeName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Scope
		wantErr bool
	}{
		{
			name:  "valid scope",
			input: "S01 Me",
			want:  &Scope{Number: 1, Name: "Me"},
		},
		{
			name:  "valid scope with multi-word name",
			input: "S02 Due Draghi",
			want:  &Scope{Number: 2, Name: "Due Draghi"},
		},
		{
			name:    "not a scope - no prefix",
			input:   "Attachments",
			wantErr: true,
		},
		{
			name:    "not a scope - dotted number",
			input:   "S01.10 Something",
			wantErr: true,
		},
		{
			name:    "not a scope - hidden dir",
			input:   ".obsidian",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseScopeName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Number != tt.want.Number || got.Name != tt.want.Name {
				t.Errorf("got {Number: %d, Name: %q}, want {Number: %d, Name: %q}",
					got.Number, got.Name, tt.want.Number, tt.want.Name)
			}
		})
	}
}

func TestParseAreaName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Area
		wantErr bool
	}{
		{
			name:  "valid area",
			input: "S01.10-19 Lifestyle",
			want:  &Area{ScopeNumber: 1, RangeStart: 10, RangeEnd: 19, Name: "Lifestyle"},
		},
		{
			name:  "valid area with multi-word name",
			input: "S02.10-19 Due Draghi al Microfono",
			want:  &Area{ScopeNumber: 2, RangeStart: 10, RangeEnd: 19, Name: "Due Draghi al Microfono"},
		},
		{
			name:  "management area",
			input: "S01.00-09 Management for S01",
			want:  &Area{ScopeNumber: 1, RangeStart: 0, RangeEnd: 9, Name: "Management for S01"},
		},
		{
			name:    "not an area - category format",
			input:   "S01.11 Entertainment",
			wantErr: true,
		},
		{
			name:    "not an area - scope format",
			input:   "S01 Me",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAreaName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ScopeNumber != tt.want.ScopeNumber || got.RangeStart != tt.want.RangeStart ||
				got.RangeEnd != tt.want.RangeEnd || got.Name != tt.want.Name {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseCategoryName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Category
		wantErr bool
	}{
		{
			name:  "valid category",
			input: "S01.11 Entertainment",
			want:  &Category{ScopeNumber: 1, Number: 11, Name: "Entertainment"},
		},
		{
			name:  "management category",
			input: "S01.10 Management for S01.10-19",
			want:  &Category{ScopeNumber: 1, Number: 10, Name: "Management for S01.10-19"},
		},
		{
			name:    "not a category - area format",
			input:   "S01.10-19 Lifestyle",
			wantErr: true,
		},
		{
			name:    "not a category - ID format",
			input:   "S01.11.11 Theatre",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCategoryName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ScopeNumber != tt.want.ScopeNumber || got.Number != tt.want.Number || got.Name != tt.want.Name {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseIDName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *ID
		wantErr bool
	}{
		{
			name:  "regular ID",
			input: "S01.11.11 Theatre, 2025 Season",
			want:  &ID{ScopeNumber: 1, CategoryNum: 11, Number: 11, Name: "Theatre, 2025 Season", IsSystemID: false},
		},
		{
			name:  "system ID - inbox",
			input: "S01.11.01 Inbox for S01.11",
			want:  &ID{ScopeNumber: 1, CategoryNum: 11, Number: 1, Name: "Inbox for S01.11", IsSystemID: true},
		},
		{
			name:  "system ID - archive",
			input: "S01.11.09 Archive for S01.11",
			want:  &ID{ScopeNumber: 1, CategoryNum: 11, Number: 9, Name: "Archive for S01.11", IsSystemID: true},
		},
		{
			name:    "not an ID - category format",
			input:   "S01.11 Entertainment",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIDName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ScopeNumber != tt.want.ScopeNumber || got.CategoryNum != tt.want.CategoryNum ||
				got.Number != tt.want.Number || got.Name != tt.want.Name || got.IsSystemID != tt.want.IsSystemID {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
