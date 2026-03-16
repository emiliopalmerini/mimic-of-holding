package vault

// Vault represents a parsed Johnny Decimal Obsidian vault.
type Vault struct {
	Root   string
	Scopes []Scope
}

// Scope represents a top-level JD scope (e.g., S01 Me).
type Scope struct {
	Number int
	Name   string
	Path   string
	Areas  []Area
}

// Area represents a JD area (e.g., S01.10-19 Lifestyle).
type Area struct {
	ScopeNumber int
	RangeStart  int
	RangeEnd    int
	Name        string
	Path        string
	Categories  []Category
}

// Category represents a JD category (e.g., S01.11 Entertainment).
type Category struct {
	ScopeNumber int
	Number      int
	Name        string
	Path        string
	IDs         []ID
}

// ID represents a JD item (e.g., S01.11.11 Theatre, 2025 Season).
type ID struct {
	ScopeNumber int
	CategoryNum int
	Number      int
	Name        string
	Path        string
	IsSystemID  bool
}
