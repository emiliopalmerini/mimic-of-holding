package main

import (
	"fmt"
	"io"

	"github.com/epalmerini/mimic-of-holding/internal/vault"
)

// scopeFromRef extracts the "S0X" prefix from any JD reference. Returns the
// empty string if the input is too short or does not start with 'S'.
func scopeFromRef(ref string) string {
	if len(ref) >= 3 && ref[0] == 'S' {
		return ref[:3]
	}
	return ""
}

// autoLog wraps vault.Log so a logging failure does not break a successful
// mutation. Errors are reported to stderr; the caller's exit code is unaffected.
func autoLog(stderr io.Writer, v *vault.Vault, scope, op, target, secondary, details string) {
	if scope == "" {
		return
	}
	if err := vault.Log(v, scope, op, target, secondary, details); err != nil {
		fmt.Fprintf(stderr, "warning: log append failed: %v\n", err)
	}
}
