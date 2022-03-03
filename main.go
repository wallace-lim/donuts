package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	stack "github.com/golang-collections/collections/stack"
)

type People struct {
	People map[int]*Person `json:"people"` // (id -> Person)
	MaxId  int             `json:"max_id"`
}

type Person struct {
	Id      int          `json:"id"`
	Name    string       `json:"name"`
	Seen    map[int]bool `json:"seen"`
	Dropped bool         `json:"dropped"`
}

type Pair struct {
	first  *Person
	second *Person
	triple bool
	third  *Person // Only valid if total is odd number
}

type State struct {
	avail_people_id []int
	matched_pairs   []*Pair
	total_people    int
}

func load_roster(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
		return nil, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
		return nil, err
	}

	peopleName := []string{}

	for _, records := range records {
		peopleName = append(peopleName, records[0])
	}

	return peopleName, nil
}

func load_state(filename string) (*People, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &People{
				People: map[int]*Person{},
				MaxId:  0,
			}, nil
		} else {
			log.Fatal("can not read file", err)
			return nil, err
		}
	}

	state := People{}
	json.Unmarshal([]byte(file), &state)
	return &state, nil
}

func persist_state(filename string, data *People) error {
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

func (p *People) update_roster(roster []string) error {
	new_people := []*Person{}

	// Identify all the new people
	for _, name := range roster {
		seen := false

		for _, person := range p.People {
			if person.Name == name {
				seen = true
				break
			}
		}

		// New person in the roster
		if !seen {
			new_people = append(new_people, &Person{
				Id:   p.MaxId + 1,
				Name: name,
				Seen: map[int]bool{},
			})

			p.MaxId++
		}
	}

	// Identify all dropped individual
	for _, person := range p.People {
		seen := false

		for _, name := range roster {
			if person.Name == name {
				seen = true
				break
			}
		}

		if !seen {
			p.People[person.Id].Dropped = true
		}
	}

	// Add all new people into list
	for _, person := range new_people {
		p.People[person.Id] = person
	}

	return nil
}

func (p *People) create_starting_state() (*State, error) {
	people_id := []int{}
	person_addr := make(map[int](*Person))

	for _, person := range p.People {
		if person.Dropped {
			continue
		}
		people_id = append(people_id, person.Id)

		person_addr[person.Id] = person
	}

	return &State{
		avail_people_id: people_id,
		matched_pairs:   []*Pair{},
		total_people:    len(p.People),
	}, nil
}

func (p *People) updateWithState(pairs []*Pair) (*People, error) {
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

	ret := make(map[int]*Person)
	for id, person := range p.People {
		ret[id] = person
	}

	return &People{
		People: ret,
		MaxId:  p.MaxId,
	}, nil
}

func (s *State) valid_pair(pair *Pair) bool {
	person_A := pair.first
	person_B := pair.second

	if person_A.Dropped || person_B.Dropped {
		return false
	}

	if _, ok := person_A.Seen[person_B.Id]; ok {
		return false
	}
	if _, ok := person_B.Seen[person_A.Id]; ok {
		return false
	}

	return true
}

func (s *State) valid_triple(pair *Pair, person_C *Person) bool {
	person_A := pair.first
	person_B := pair.second

	if person_C.Dropped {
		return false
	}
	if _, ok := person_A.Seen[person_C.Id]; ok {
		return false
	}
	if _, ok := person_B.Seen[person_C.Id]; ok {
		return false
	}

	return true
}

func (s *State) get_successors(people *People) ([]*State, error) {
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

			// Verify pair has not seen each other before
			if !s.valid_pair(pair) {
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

	// Loop until empty
	for fringe.Len() > 0 {
		// Remove node from fringe
		curr_state := fringe.Pop().(*State)

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
			fringe.Push(new_state)
		}
	}
	return nil, errors.New("no possible matches")
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
			pair.first.Name, pair.second.Name, "",
		}

		if pair.triple {
			pair_str[2] = pair.third.Name
		}

		// Debugging purposes
		// fmt.Println(pair_str)

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
	rosterFilePath := "roster.csv"

	people, err := load_state(in_file)
	if err != nil {
		log.Fatal(err)
		return
	}

	curr_roster, err := load_roster(rosterFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = people.update_roster(curr_roster)
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
	err = persist_state("seen.json", newBlob)
	if err != nil {
		log.Fatal(err)
		return
	}
}
