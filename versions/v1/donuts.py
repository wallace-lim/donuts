import numpy as np
import pandas as pd
import csv
import argparse

def generate_pairs(arr):
    """
    Create a list of pairings given an arr. The pairing
    matches each element on both ends together and pairs inwards.
    
    If there's an odd number of people, the middle element will
    join the first pairing to be a triple
    
    @params arr (arr) - list of people to pair up
    
    @return (list) - list of tuples signifying the pairings
    """
    
    pairs = []
    
    for i in range(len(arr)//2):
        pairs.append((arr[i], arr[-i-1]))
    
    if len(arr) % 2 == 1:
        pairs[0] = (*pairs[0], arr[len(arr)//2])
    return pairs

def generate_matches(n, t):
    """
    Generate t pairing possibilities of n individuals using
    a round-robin algorithm.
    
    Algorithm Resourse
    See: https://stackoverflow.com/questions/54447564/an-efficient-approach-to-combinations-of-pairs-in-groups-without-repetitions
    Alternative Solution
    Alt: https://math.stackexchange.com/questions/3093225/an-efficient-approach-to-combinations-of-pairs-in-groups-without-repetitions
    
    @params n (int) - number of people in matching process
    @params t (int) - number of times to pair entire group up
    
    @return (list) - list of pairings
    """
    
    arr = np.arange(n)
    rotate_idx = np.hstack(([0], np.roll(np.arange(1,n), shift=1)))
    
    matches = []
    for _ in range(t):
        arr = arr[rotate_idx]
        matches.append(generate_pairs(arr))
    return matches

def double_check_valid(n, matches):
    """
    Returns True if there are no duplicate pairing in all generated matches
    
    @params n (int) - number of people in matching process
    @params matches (list) - list of all pairing for each matching process
    
    @return (bool) - T if no duplicates, False otherwise
    """
    visited = {i:set([i]) for i in range(n)}
    
    for match in matches:
        for pair in match:
            pair_set = set(pair)
            for person in pair:
                # Find intersection between who person has visited and current pair
                dupes = visited[person].intersection(pair_set)
                if len(dupes) > 1:
                    return False
                
                # Add into set all people in pair
                visited[person] = visited[person].union(pair_set)
            
    return True

def donuts(in_filepath, out_filepath, num_meets, seed=1):
    # Set a random seed to shuffle
    np.random.seed(seed)

    # Read in name listing
    name_lst = pd.read_csv(in_filepath, names=['Names']).squeeze()
    name_lst = np.random.permutation(name_lst.values)

    n = len(name_lst)

    # Perform Pair Matching
    matches = generate_matches(n, num_meets)

    if not double_check_valid(n, matches):
        print("Contains duplicates: lower the number of meets")
        print("Aborting...")
        return

    # Generate CSV File
    headers = ['Match', 'Person 1', 'Person 2', 'Person 3']
    df = pd.DataFrame(columns=headers)

    # Write Each Match's Pairings
    for i in range(num_meets):
        match = matches[i]
        for pair in match:
            pair_names = [name_lst[j] for j in pair]
            
            # Add row to dataframe
            df.loc[df.shape[0]] = [i] + pair_names + ([None] if len(pair) == 2 else [])

    df.to_csv('pairing.csv', index=False)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--in_file", "-i", type=str, default='names.csv', help="Specify file location of list of names (csv file)")
    parser.add_argument("--out_file", "-o", type=str, default='pairings.csv', help="Specify file location of output file (csv file)")
    parser.add_argument("--num_meets", "-m", type=int, help="Specify number of desired meetings")
    parser.add_argument("--seed", "-s", type=int, help="Specify a seed for initial shuffling")
    args = parser.parse_args()
    print(args.seed)
    donuts(args.in_file, args.out_file, args.num_meets, args.seed)

