// Feature-flag / experiment cohort example.
//
// A product typically maintains several overlapping cohorts of users:
//
//   - "beta"        - users who opted into the beta program
//   - "paid"        - users on a paid plan
//   - "experiment"  - 10% of users bucketed into a new A/B experiment
//
// Gating logic then asks questions like:
//
//   - is this user in ANY of these cohorts?
//   - is this user in ALL of these cohorts?
//   - how large is the intersection "paid AND experiment"?
//
// This example also demonstrates the .bin persistence layer: the cohort
// state is saved to disk, reloaded in a fresh process, and used to answer
// a few gating queries.
package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"

	"github.com/rschoonheim/gg"
)

const userCount = 10_000

func main() {
	dir, err := os.MkdirTemp("", "gg-cohorts-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "cohorts.bin")

	build(path)
	query(path)
}

// build populates three cohorts with deterministic PRNG draws and saves
// them to a `.bin` file on disk.
func build(path string) {
	gs := gg.New(userCount)
	r := rand.New(rand.NewPCG(1, 2))

	beta := pick(r, userCount, 0.20)       // 20% opted into beta
	paid := pick(r, userCount, 0.05)       // 5% on a paid plan
	experiment := pick(r, userCount, 0.10) // 10% in the A/B bucket

	if _, err := gs.Add("beta", beta...); err != nil {
		panic(err)
	}
	if _, err := gs.Add("paid", paid...); err != nil {
		panic(err)
	}
	if _, err := gs.Add("experiment", experiment...); err != nil {
		panic(err)
	}

	if err := gs.SaveFile(path); err != nil {
		panic(err)
	}
	info, _ := os.Stat(path)
	fmt.Printf("saved %s (%d bytes) with %d cohorts over %d users\n",
		path, info.Size(), gs.Len(), gs.EntityCount())
}

// query reloads the cohorts file and runs a few gating checks.
func query(path string) {
	gs, err := gg.LoadFile(path)
	if err != nil {
		panic(err)
	}

	beta, _ := gs.Get("beta")
	paid, _ := gs.Get("paid")
	experiment, _ := gs.Get("experiment")

	fmt.Println("\n== cohort sizes ==")
	fmt.Printf("  beta      : %d\n", beta.Cardinality())
	fmt.Printf("  paid      : %d\n", paid.Cardinality())
	fmt.Printf("  experiment: %d\n", experiment.Cardinality())

	// Reach of "paid AND experiment" — how many users experience both.
	paidExp, _ := paid.Intersection(experiment)
	fmt.Printf("\npaid ∩ experiment: %d users\n", paidExp.Cardinality())

	// Any of the three (gated feature surface).
	anyCohort, _ := beta.Union(paid)
	anyCohort, _ = anyCohort.Union(experiment)
	fmt.Printf("beta ∪ paid ∪ experiment: %d users\n", anyCohort.Cardinality())

	// Per-user gating using the reverse-lookup helpers.
	for _, user := range []uint32{0, 42, 1234, 9999} {
		cohorts := gs.FindNames(user)
		fmt.Printf("user %-4d belongs to %v\n", user, cohorts)
	}

	// Audience targeting: users who are paid AND in the experiment but NOT
	// in beta (e.g. to avoid double-exposure during a rollout).
	target, _ := paidExp.Difference(beta)
	fmt.Printf("\ntarget audience (paid ∩ experiment) \\ beta: %d users\n", target.Cardinality())
}

// pick draws a deterministic pseudo-random subset of [0, universe) at the
// requested fill ratio.
func pick(r *rand.Rand, universe uint32, ratio float64) []uint32 {
	out := make([]uint32, 0, int(float64(universe)*ratio)+1)
	for i := uint32(0); i < universe; i++ {
		if r.Float64() < ratio {
			out = append(out, i)
		}
	}
	return out
}

