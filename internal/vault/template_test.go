package vault

import "testing"

// --- ApplyTemplate unit tests ---

func TestApplyTemplate_AllVarsSubstituted(t *testing.T) {
	content := "ref={{ref}} name={{name}} title={{title}} date={{date}}"
	vars := TemplateVars{
		Ref:   "S01.11.12",
		Name:  "Cinema",
		Title: "S01.11.12 Cinema",
		Date:  "2026-03-26",
	}
	got := ApplyTemplate(content, vars)
	want := "ref=S01.11.12 name=Cinema title=S01.11.12 Cinema date=2026-03-26"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApplyTemplate_UnknownVarsLeftAlone(t *testing.T) {
	content := "{{ref}} and {{porzioni}} and {{calorie}}"
	vars := TemplateVars{Ref: "S01.11.12"}
	got := ApplyTemplate(content, vars)
	want := "S01.11.12 and {{porzioni}} and {{calorie}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApplyTemplate_NoVars(t *testing.T) {
	content := "No variables here."
	got := ApplyTemplate(content, TemplateVars{})
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestApplyTemplate_CustomVarsSubstituted(t *testing.T) {
	content := "Rating: {{rating}} Venue: {{venue}}"
	vars := TemplateVars{
		CustomVars: map[string]string{
			"rating": "5/5",
			"venue":  "Teatro Regio",
		},
	}
	got := ApplyTemplate(content, vars)
	want := "Rating: 5/5 Venue: Teatro Regio"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestApplyTemplate_BuiltinPrecedenceOverCustom(t *testing.T) {
	content := "{{ref}} {{name}}"
	vars := TemplateVars{
		Ref:  "S01.11.11",
		Name: "Theatre",
		CustomVars: map[string]string{
			"ref":  "CUSTOM_REF",
			"name": "CUSTOM_NAME",
		},
	}
	got := ApplyTemplate(content, vars)
	want := "S01.11.11 Theatre"
	if got != want {
		t.Errorf("built-in should win: got %q, want %q", got, want)
	}
}

func TestApplyTemplate_EmptyCustomVarsNoChange(t *testing.T) {
	content := "{{rating}} and {{venue}}"
	vars := TemplateVars{CustomVars: map[string]string{}}
	got := ApplyTemplate(content, vars)
	if got != content {
		t.Errorf("empty custom vars should not change content: got %q", got)
	}
}

func TestApplyTemplate_NilCustomVarsNoChange(t *testing.T) {
	content := "{{rating}} and {{venue}}"
	vars := TemplateVars{}
	got := ApplyTemplate(content, vars)
	if got != content {
		t.Errorf("nil custom vars should not change content: got %q", got)
	}
}

func TestApplyTemplate_CustomVarEmptyStringValue(t *testing.T) {
	content := "Rating: {{rating}}"
	vars := TemplateVars{
		CustomVars: map[string]string{"rating": ""},
	}
	got := ApplyTemplate(content, vars)
	want := "Rating: "
	if got != want {
		t.Errorf("empty value should replace placeholder: got %q, want %q", got, want)
	}
}

func TestApplyTemplate_CustomVarNotInTemplate(t *testing.T) {
	content := "No placeholders here."
	vars := TemplateVars{
		CustomVars: map[string]string{"rating": "5/5"},
	}
	got := ApplyTemplate(content, vars)
	if got != content {
		t.Errorf("unknown custom var should not change content: got %q", got)
	}
}

// --- ListTemplates unit tests ---

func TestListTemplates_InvalidRef(t *testing.T) {
	_, err := ListTemplates(searchFixture, "xyz")
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestListTemplates_NonCategoryRef(t *testing.T) {
	_, err := ListTemplates(searchFixture, "S01")
	if err == nil {
		t.Fatal("expected error for scope ref")
	}
	_, err = ListTemplates(searchFixture, "S01.10-19")
	if err == nil {
		t.Fatal("expected error for area ref")
	}
	_, err = ListTemplates(searchFixture, "S01.11.11")
	if err == nil {
		t.Fatal("expected error for ID ref")
	}
}
