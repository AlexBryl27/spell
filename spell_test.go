package spell

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/eskriett/strmet"
)

func BenchmarkSpell_Lookup(b *testing.B) {
	s, err := newWithExample()
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		s.Lookup("exampl")
	}
}

func ExampleSpell_AddEntry() {
	// Create a new speller
	s := New()

	// Add a new word, "example" to the dictionary
	s.AddEntry(Entry{
		Frequency: 10,
		Word:      "example",
	})

	// Overwrite the data for word "example"
	s.AddEntry(Entry{
		Frequency: 100,
		Word:      "example",
	})

	// Output the frequency for word "example"
	entry := s.GetEntry("example")
	fmt.Printf("Output for word 'example' is: %v\n",
		entry.Frequency)
	// Output:
	// Output for word 'example' is: 100
}

func ExampleSpell_Lookup() {
	// Create a new speller
	s := New()
	s.AddEntry(Entry{
		Frequency: 1,
		Word:      "example",
	})

	// Perform a default lookup for example
	suggestions, _ := s.Lookup("eample")
	fmt.Printf("Suggestions are: %v\n", suggestions)
	// Output:
	// Suggestions are: [example]
}

func ExampleSpell_Lookup_configureEditDistance() {
	// Create a new speller
	s := New()
	s.AddEntry(Entry{
		Frequency: 1,
		Word:      "example",
	})

	// Lookup exact matches, i.e. edit distance = 0
	suggestions, _ := s.Lookup("eample", EditDistance(0))
	fmt.Printf("Suggestions are: %v\n", suggestions)
	// Output:
	// Suggestions are: []
}

func ExampleSpell_Lookup_configureDistanceFunc() {
	// Create a new speller
	s := New()
	s.AddEntry(Entry{
		Frequency: 1,
		Word:      "example",
	})

	// Configure the Lookup to use Levenshtein distance rather than the default
	// Damerau Levenshtein calculation
	s.Lookup("example", DistanceFunc(func(s1, s2 string, maxDist int) int {
		// Call the Levenshtein function from github.com/eskriett/strmet
		return strmet.Levenshtein(s1, s2, maxDist)
	}))
}

func ExampleSpell_Lookup_configureSortFunc() {
	// Create a new speller
	s := New()
	s.AddEntry(Entry{
		Frequency: 1,
		Word:      "example",
	})

	// Configure suggestions to be sorted solely by their frequency
	s.Lookup("example", SortFunc(func(sl SuggestionList) {
		sort.Slice(sl, func(i, j int) bool {
			return sl[i].Frequency < sl[j].Frequency
		})
	}))
}

func ExampleSpell_Segment() {
	// Create a new speller
	s := New()

	s.AddEntry(Entry{Frequency: 1, Word: "the"})
	s.AddEntry(Entry{Frequency: 1, Word: "quick"})
	s.AddEntry(Entry{Frequency: 1, Word: "brown"})
	s.AddEntry(Entry{Frequency: 1, Word: "fox"})

	// Segment a string with word concatenated together
	segmentResult, _ := s.Segment("thequickbrownfox")
	fmt.Println(segmentResult)
	// Output:
	// the quick brown fox
}

func newWithExample() (*Spell, error) {
	s := New()
	ok, err := s.AddEntry(Entry{
		Frequency: 1,
		Word:      "example",
	})
	if err != nil {
		return s, err
	}
	if !ok {
		return s, errors.New("failed to insert entry")
	}
	return s, nil
}

func TestAddEntry(t *testing.T) {
	_, err := newWithExample()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLookup(t *testing.T) {
	s, err := newWithExample()
	if err != nil {
		t.Fatal(err)
	}

	suggestions, err := s.Lookup("eample")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 {
		t.Fatal("did not get exactly one match")
	}
	if suggestions[0].Word != "example" {
		t.Fatal(fmt.Sprintf("Expected example, got %s", suggestions[0].Word))
	}

	// Test Unicode strings
	suggestions, err = s.Lookup("ex𝐚mple")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 {
		t.Fatal("did not get exactly one match")
	}
	if suggestions[0].Word != "example" {
		t.Fatal(fmt.Sprintf("Expected example, got %s", suggestions[0].Word))
	}

	ok, err := s.AddEntry(Entry{
		Frequency: 1,
		Word:      "ex𝐚mple",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Failed to add unicode entry")
	}

	suggestions, err = s.Lookup("ex𝐚mple")
	if err != nil {
		t.Fatal(err)
	}
	if suggestions[0].Word != "ex𝐚mple" {
		t.Fatal(fmt.Sprintf("Expected ex𝐚mple, got %s", suggestions[0].Word))
	}
}

func TestRemoveEntry(t *testing.T) {
	s, err := newWithExample()
	if err != nil {
		t.Fatal(err)
	}
	if ok := s.RemoveEntry("example"); !ok {
		t.Fatal("failed to remove entry")
	}
	suggestions, err := s.Lookup("example")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 0 {
		t.Fatal("did not get exactly zero matches")
	}
	if ok := s.RemoveEntry("example"); ok {
		t.Fatal("should not remove twice")
	}
}

func TestLongestWord(t *testing.T) {
	s, err := newWithExample()
	if err != nil {
		t.Fatal(err)
	}
	if wordLength := s.GetLongestWord(); wordLength != uint32(len("example")) {
		t.Fatal("failed to get longest word length, expected 7 got: ", wordLength)
	}
}

func TestSaveLoad(t *testing.T) {
	s1, err := newWithExample()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./test.dump")
	if err := s1.Save("./test.dump"); err != nil {
		t.Fatal(err)
	}
	s2, err := Load("./test.dump")
	if err != nil {
		t.Fatal(err)
	}
	suggestions, err := s2.Lookup("eample")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 {
		t.Fatal("did not get exactly one match")
	}
	if suggestions[0].Word != "example" {
		t.Fatal(fmt.Sprintf("Expected example, got %s", suggestions[0].Word))
	}
}

func TestCornerCases(t *testing.T) {
	s := New()
	ok, err := s.AddEntry(Entry{
		Frequency: 1,
		Word:      "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("failed to add entry to speller")
	}
	suggestions, err := s.Lookup("a")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggestions) != 1 {
		t.Fatal("did not get exactly one match")
	}
	if suggestions[0].Word != "" {
		t.Fatal(fmt.Sprintf("Expected ' ', got %s", suggestions[0].Word))
	}
}
