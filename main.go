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

func load_family(filePath string) ([][]string, error) {
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

	familes := [][]string{}

	for _, records := range records {
		family := append([]string{}, records...)
		familes = append(familes, family)
	}

	return familes, nil
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

	directory := "data/"
	persistFilePath := directory + "seen.json"
	csv_filename := directory + "pairing.csv"
	rosterFilePath := directory + "roster.csv"
	familyFilePath := directory + "family.csv"

	people, err := load_state(persistFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	curr_roster, err := load_roster(rosterFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	family, err := load_family(familyFilePath)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(family)

	err = people.update_roster(curr_roster)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = people.mark_family(family)
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
	err = persist_state(persistFilePath, newBlob)
	if err != nil {
		log.Fatal(err)
		return
	}
}
