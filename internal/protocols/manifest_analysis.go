package protocols

import (
	"encoding/json"
	"sort"
)

// PermissionDiff is an offline review artifact; it never grants permissions.
type PermissionDiff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
	Changed bool     `json:"changed"`
}

// Analysis shares the authoring schema; readable guides stay inert and no
// remote operation is accepted. Publication separately validates licensing.
func AnalyzeManifest(raw []byte) error {
	if ValidatePreviewPermissions(raw) != nil {
		return ErrUnsafePermissions
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil || body == nil {
		return ErrUnsafePermissions
	}
	if _, ok := body["steps"]; !ok {
		body["steps"] = json.RawMessage(`[]`)
	}
	normalized, err := json.Marshal(body)
	if err != nil || ValidateDraftSpecWithGuides(normalized, "MIT") != nil {
		return ErrUnsafePermissions
	}
	return nil
}
func DiffPermissions(before, after []string) PermissionDiff {
	a, b := map[string]bool{}, map[string]bool{}
	for _, x := range before {
		a[x] = true
	}
	for _, x := range after {
		b[x] = true
	}
	diff := PermissionDiff{Added: []string{}, Removed: []string{}}
	for x := range b {
		if !a[x] {
			diff.Added = append(diff.Added, x)
		}
	}
	for x := range a {
		if !b[x] {
			diff.Removed = append(diff.Removed, x)
		}
	}
	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)
	diff.Changed = len(diff.Added)+len(diff.Removed) > 0
	return diff
}
