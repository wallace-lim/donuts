package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	stack "github.com/golang-collections/collections/stack"
)

type People struct {
	People map[int]Person `json:"people"` // (id -> Person)
}

type Person struct {
	Id   int          `json:"id"`
	Name string       `json:"name"`
	Seen map[int]bool `json:"seen"`
}

type Pair struct {
	first  Person
	second Person
	triple bool
	third  Person // Only valid if total is odd number
}

type State struct {
	avail_people_id []int
	matched_pairs   []*Pair
	total_people    int
}

func load_file(filename string) (*People, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("can not read file", err)
		return nil, err
	}

	state := People{}
	json.Unmarshal([]byte(file), &state)
	return &state, nil
}

func persist_file(filename string, data *People) error {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal("can not marshal json", err)
		return err
	}

	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		log.Fatal("can not write file", err)
		return err
	}

	return nil
}

func (p People) create_starting_state() (*State, error) {
	people_id := []int{}
	person_addr := make(map[int](*Person))

	for _, person := range p.People {
		people_id = append(people_id, person.Id)

		person_addr[person.Id] = &person
	}

	return &State{
		avail_people_id: people_id,
		matched_pairs:   []*Pair{},
		total_people:    len(p.People),
	}, nil
}

func (s State) valid_pair(pair *Pair) bool {
	person_A := pair.first
	person_B := pair.second

	if _, ok := person_A.Seen[person_B.Id]; ok {
		return false
	}
	if _, ok := person_B.Seen[person_A.Id]; ok {
		return false
	}

	return true
}

func (s State) valid_triple(pair *Pair, person_C Person) bool {
	person_A := pair.first
	person_B := pair.second

	if _, ok := person_A.Seen[person_C.Id]; ok {
		return false
	}
	if _, ok := person_B.Seen[person_C.Id]; ok {
		return false
	}

	return true
}

func (s State) get_successors(people *People) ([]*State, error) {
	childrenState := []*State{}

	// Handle odd numder of people case
	if len(s.avail_people_id) == 1 {
		person_C := people.People[s.avail_people_id[0]]

		for _, pair := range s.matched_pairs {

			if !s.valid_triple(pair, person_C) {
				continue
			}

			// Add triple into pair
			pair.third = person_C
			pair.triple = true

			newState := &State{
				avail_people_id: []int{},
				matched_pairs:   s.matched_pairs,
				total_people:    s.total_people,
			}

			return append(childrenState, newState), nil
		}

		return childrenState, nil
	}

	for i := 0; i < len(s.avail_people_id); i++ {
		for j := i + 1; j < len(s.avail_people_id); j++ {
			person_A := people.People[s.avail_people_id[i]]
			person_B := people.People[s.avail_people_id[j]]

			pair := &Pair{
				first:  person_A,
				second: person_B,
			}

			// fmt.Printf("i: %d, j: %d\n", i, j)
			// fmt.Println(s.avail_people_id)
			// fmt.Printf("Matched person A: %d with person B: %d\n", person_A.Id, person_B.Id)

			// Verify pair has not seen each other before
			if !s.valid_pair(pair) {
				// fmt.Println("seen before")
				continue
			}

			// Create new state
			var remainder_people_id []int = []int{}
			for _, id := range s.avail_people_id {
				if id != person_A.Id && id != person_B.Id {
					remainder_people_id = append(remainder_people_id, id)
				}
			}

			newState := &State{
				avail_people_id: remainder_people_id,
				matched_pairs:   append(s.matched_pairs, pair),
				total_people:    s.total_people,
			}

			// Append to possible childrenState
			childrenState = append(childrenState, newState)
		}
	}

	return childrenState, nil
}

func run(state *State, people *People) ([]*Pair, error) {
	fringe := stack.New()
	// Insert current state
	fringe.Push(state)

	// tmp := 0

	// Loop until empty
	for fringe.Len() > 0 {
		// Remove node from fringe
		curr_state := fringe.Pop().(*State)

		// fmt.Println("-------------------------------")
		// fmt.Println("Curr State")
		// fmt.Println(curr_state.avail_people_id)
		// fmt.Println(curr_state.matched_pairs)

		// Verify goal state
		if len(curr_state.avail_people_id) == 0 {
			return curr_state.matched_pairs, nil
		}

		// Find neighbors
		next_states, err := curr_state.get_successors(people)
		if err != nil {
			return nil, err
		}

		for _, new_state := range next_states {
			// fmt.Println("Child State")
			// fmt.Println(new_state.avail_people_id)
			// fmt.Println(new_state.matched_pairs[0].first)
			// fmt.Println(new_state.matched_pairs[0].second)

			fringe.Push(new_state)
		}

		// tmp += 1

		// if tmp > 0 {
		// 	return nil, nil
		// }
	}
	return nil, errors.New("no possible matches")
}

func (p People) updateWithState(pairs []*Pair) (*People, error) {
	// Iterate over all pairs
	for _, pair := range pairs {
		// Add each pair to seen
		person_A := p.People[pair.first.Id]
		person_B := p.People[pair.second.Id]

		person_A.Seen[person_B.Id] = true
		person_B.Seen[person_A.Id] = true

		// Hand odd case
		if pair.triple {
			person_C := p.People[pair.third.Id]

			person_A.Seen[person_C.Id] = true
			person_B.Seen[person_C.Id] = true

			person_C.Seen[person_A.Id] = true
			person_C.Seen[person_B.Id] = true
		}
	}

	ret := make(map[int]Person)
	for id, person := range p.People {
		ret[id] = person
	}

	return &People{
		People: ret,
	}, nil
}

func pairsToCSV(filename string, pairs []*Pair) error {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalln("failed to open file", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, pair := range pairs {
		pair_str := []string{
			pair.first.Name, pair.second.Name, pair.third.Name,
		}
		fmt.Println(pair_str)

		if err := w.Write(pair_str); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	return nil
}

func main() {
	// args := os.Args[1:]
	// in_file := args[1]
	// out_file := args[2]

	in_file := "seen.json"
	csv_filename := "pairing.csv"

	people, err := load_file(in_file)
	if err != nil {
		log.Fatal(err)
		return
	}

	state, err := people.create_starting_state()
	if err != nil {
		log.Fatal(err)
		return
	}

	pairs, err := run(state, people)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Write result fixing pairs
	err = pairsToCSV(csv_filename, pairs)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Update people
	newBlob, err := people.updateWithState(pairs)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Persist file
	err = persist_file("seen.json", newBlob)
	if err != nil {
		log.Fatal(err)
		return
	}
}
