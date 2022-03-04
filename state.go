package main

type State struct {
	avail_people_id []int
	matched_pairs   []*Pair
	total_people    int
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
