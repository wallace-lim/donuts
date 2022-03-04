# Donuts

## Table of Contents
  - [About](#about)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installing](#installing)
    - [Requirements](#requirements)
    - [Running Application](#running-application)
    - [Future Runs](#future-runs)

## About
WIP

## Getting Started

### Prerequisites
In order to proper run the code, you must have [GoLang](https://go.dev/doc/install) installed

### Installing
First, clone the package and retrieve all the dependencies

```
git clone https://github.com/wallace-lim/donuts

go get -d ./...
```

### Requirements
To properly generate valid pairings, there are a couple formatting conditions for this to properly work. 

Create a folder (default use `data/`) that will contain the following contents
- `roster.csv` - a single column csv where a row contains a name in the current roster
- `family.csv` - a file where a row contains the names of a family group (aka. a group that should not have pairing between each other)

Check [sample_data/](https://github.com/wallace-lim/donuts/tree/main/sample_data) for how the data should be formatted
- [roster.csv](https://raw.githubusercontent.com/wallace-lim/donuts/main/sample_data/roster.csv)
- [family.csv](https://raw.githubusercontent.com/wallace-lim/donuts/main/sample_data/family.csv)

### Running Application
Finally, we can run the application!

```
go run *.go <directoryPath>
go run *.go sample_data
```
If `<directoryPath>` is not specified, the program by default will expect the donut contents to be in `data/`

After a successful run, you should expect a few files to appear/by modified in the folder
- `pairing.csv` - a file where every row is a pairing outputted by the algorithm
- `seen.json` - persisted metadata for future runs of algorithm (feel free do ignore contents)

### Future Runs
After the initial generation, we can run the same command as many times as we want to continuously generate pairings as long as it is possible to create a new valid pairing set. `seen.json` will maintain a persisted state to keep track of which individual has seen who, and the output will be located in `pairing.csv`

```
go run *.go <directoryPath>
```
