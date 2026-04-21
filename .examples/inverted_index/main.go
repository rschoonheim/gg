// Inverted-index example.
//
// Each document is represented by an integer id. For every indexed term,
// a Grouping holds the set of document-ids that contain that term. A
// keyword search is then a straightforward boolean combination of
// bitmaps:
//
//   - "foo AND bar"     -> intersection
//   - "foo OR bar"      -> union
//   - "foo AND NOT bar" -> difference
//
// This is exactly how large-scale search engines (Lucene, Tantivy, …)
// implement their posting-list boolean layer, albeit with much more
// sophisticated compression.
package main

import (
	"fmt"
	"strings"

	"gg"
)

var corpus = []string{
	"The quick brown fox jumps over the lazy dog",
	"Pack my box with five dozen liquor jugs",
	"The five boxing wizards jump quickly",
	"How vexingly quick daft zebras jump",
	"Sphinx of black quartz judge my vow",
	"A quick brown dog outpaces a quick fox",
}

func main() {
	// Tokenise and build the inverted index.
	terms := map[string][]uint32{}
	for docID, doc := range corpus {
		seen := map[string]bool{}
		for _, tok := range strings.Fields(strings.ToLower(doc)) {
			tok = strings.Trim(tok, ".,;:!?")
			if tok == "" || seen[tok] {
				continue
			}
			seen[tok] = true
			terms[tok] = append(terms[tok], uint32(docID))
		}
	}

	idx := gg.New(uint32(len(corpus)))
	for term, docs := range terms {
		if _, err := idx.Add(term, docs...); err != nil {
			panic(err)
		}
	}

	// Helper: resolve a term to its posting-list grouping (or the empty
	// grouping if the term is unknown).
	posting := func(term string) *gg.Grouping {
		if g, ok := idx.Get(term); ok {
			return g
		}
		empty, _ := gg.New(uint32(len(corpus))).Add("∅")
		return empty
	}

	// Run a few queries.
	runQuery := func(label string, result *gg.Grouping) {
		fmt.Printf("\n== %s (%d hits) ==\n", label, result.Cardinality())
		for _, id := range result.Members() {
			fmt.Printf("  [%d] %s\n", id, corpus[id])
		}
	}

	quick := posting("quick")
	brown := posting("brown")
	jump := posting("jump")

	and, _ := quick.Intersection(brown)
	or, _ := quick.Union(jump)
	andNot, _ := quick.Difference(brown)

	runQuery(`"quick" AND "brown"`, and)
	runQuery(`"quick" OR "jump"`, or)
	runQuery(`"quick" AND NOT "brown"`, andNot)

	// Reverse lookup: "which terms index document 5?".
	fmt.Println("\n== terms indexing document 5 ==")
	fmt.Println(idx.FindNames(5))
}

