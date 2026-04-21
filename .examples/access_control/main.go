// Access-control example.
//
// Each user is represented by an integer id in [0, N). A Grouping models
// a role (admin, editor, viewer, …) as the set of user-ids holding that
// role. Answering real-world questions boils down to set-theoretic
// operations on bitmaps:
//
//   - "Which users are both editors and reviewers?" -> intersection
//   - "Which users have any of these roles?"        -> union
//   - "Which users are viewers but not editors?"    -> difference
//   - "Which roles does user 42 hold?"              -> reverse lookup
package main

import (
	"fmt"

	"gg"
)

// userNames maps user ids to human-readable names. In a real system this
// would come from a database; here it's just for pretty-printing.
var userNames = []string{
	0: "alice", 1: "bob", 2: "carol", 3: "dan",
	4: "eve", 5: "frank", 6: "grace", 7: "heidi",
}

func main() {
	const population = 8
	gs := gg.New(population)

	// Populate roles.
	mustAdd(gs, "admin", 0)                   // alice
	mustAdd(gs, "editor", 0, 2, 4)            // alice, carol, eve
	mustAdd(gs, "reviewer", 2, 3, 4, 6)       // carol, dan, eve, grace
	mustAdd(gs, "viewer", 1, 2, 3, 4, 5, 6, 7)

	editor, _ := gs.Get("editor")
	reviewer, _ := gs.Get("reviewer")
	viewer, _ := gs.Get("viewer")

	// Forward queries: bitwise set operations on bitmaps.
	editorsAndReviewers, _ := editor.Intersection(reviewer)
	anyPrivileged, _ := editor.Union(reviewer)
	viewersOnly, _ := viewer.Difference(anyPrivileged)

	fmt.Println("== forward queries ==")
	fmt.Println("editors ∩ reviewers :", named(editorsAndReviewers.Members()))
	fmt.Println("editors ∪ reviewers :", named(anyPrivileged.Members()))
	fmt.Println("viewers \\ privileged:", named(viewersOnly.Members()))

	// Reverse lookup: "what roles does carol (id 2) hold?".
	fmt.Println("\n== reverse lookup ==")
	for _, r := range gs.FindNames(2) {
		fmt.Println("carol is", r)
	}

	// Authorisation check.
	fmt.Println("\n== authorisation ==")
	fmt.Println("can alice edit?", editor.Contains(0))
	fmt.Println("can bob edit? ", editor.Contains(1))

	// Invariant checks expressed with set theory.
	fmt.Println("\n== invariants ==")
	admin, _ := gs.Get("admin")
	fmt.Println("admins ⊆ editors :", admin.IsSubsetOf(editor))
}

func mustAdd(gs *gg.Groupings, name string, members ...uint32) {
	if _, err := gs.Add(name, members...); err != nil {
		panic(err)
	}
}

func named(ids []uint32) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = userNames[id]
	}
	return out
}

