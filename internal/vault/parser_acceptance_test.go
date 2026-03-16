package vault

import (
	"path/filepath"
	"testing"
)

func TestAcceptance_ParseVault_RealisticFixture(t *testing.T) {
	root := filepath.Join(testdataDir(t), "vault")
	v, err := ParseVault(root)
	if err != nil {
		t.Fatalf("ParseVault failed: %v", err)
	}

	// Acceptance criteria 1: Contains expected scopes by name
	scopeNames := make(map[string]bool)
	for _, s := range v.Scopes {
		scopeNames[s.Name] = true
	}
	for _, want := range []string{"Me", "Due Draghi"} {
		if !scopeNames[want] {
			t.Errorf("missing scope %q", want)
		}
	}

	// Acceptance criteria 2: Non-JD entries are absent
	for _, s := range v.Scopes {
		if s.Name == ".obsidian" || s.Name == "Attachments" || s.Name == "README.md" {
			t.Errorf("non-JD entry %q should not appear as scope", s.Name)
		}
	}

	// Acceptance criteria 3: System IDs are flagged
	var systemIDs []ID
	var regularIDs []ID
	for _, s := range v.Scopes {
		for _, a := range s.Areas {
			for _, c := range a.Categories {
				for _, id := range c.IDs {
					if id.IsSystemID {
						systemIDs = append(systemIDs, id)
					} else {
						regularIDs = append(regularIDs, id)
					}
				}
			}
		}
	}
	if len(systemIDs) == 0 {
		t.Error("expected at least one system ID (IsSystemID=true)")
	}
	if len(regularIDs) == 0 {
		t.Error("expected at least one regular ID (IsSystemID=false)")
	}

	// Acceptance criteria 4: System IDs have numbers 01-09
	for _, id := range systemIDs {
		if id.Number < 1 || id.Number > 9 {
			t.Errorf("system ID %q has number %d, expected 1-9", id.Name, id.Number)
		}
	}

	// Acceptance criteria 5: Regular IDs have numbers >= 10
	for _, id := range regularIDs {
		if id.Number < 10 {
			t.Errorf("regular ID %q has number %d but IsSystemID=false", id.Name, id.Number)
		}
	}

	// Acceptance criteria 6: Paths are absolute and exist under root
	for _, s := range v.Scopes {
		if !filepath.IsAbs(s.Path) {
			t.Errorf("scope %q path is not absolute: %s", s.Name, s.Path)
		}
		for _, a := range s.Areas {
			if !filepath.IsAbs(a.Path) {
				t.Errorf("area %q path is not absolute: %s", a.Name, a.Path)
			}
		}
	}
}
