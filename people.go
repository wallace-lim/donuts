package main

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
