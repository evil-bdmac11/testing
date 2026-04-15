# Battleship (Go CLI)

This repository now includes a complete command-line implementation of **Battleship** in Go.

## Location

All game code lives in:

- `./battleship/`

## Features

- Standard 10x10 Battleship boards (A-J, 1-10)
- Full standard fleet:
  - Carrier (5)
  - Battleship (4)
  - Cruiser (3)
  - Submarine (3)
  - Destroyer (2)
- Manual player ship placement with validation
- Random valid computer ship placement
- Turn-based firing, hit/miss/sunk detection
- Win condition when one fleet is fully destroyed
- Smarter computer AI:
  - Hunts randomly at first
  - Targets adjacent cells after hits
  - Follows hit lines to sink ships
  - Never repeats shots
- Side-by-side boards during gameplay
- ANSI terminal colors for water, ships, hits, misses, and alerts
- ASCII title and ship art
- Hit/miss dramatic reveal and sunk announcements
- `help` command during player turn
- End-game battle summary with accuracy statistics

## Build

From repository root:

```bash
go build -o ./battleship/battleship ./battleship/
```

Or from inside the module:

```bash
cd battleship
go build
```

## Run

From inside the module after building:

```bash
cd battleship
./battleship
```

## Controls

- Enter coordinates like `A5`, `J10`
- During ship placement, choose orientation:
  - `H` = horizontal
  - `V` = vertical
- During your turn:
  - Enter `help` to view instructions

## Example Gameplay (text snapshot)

```text
YOUR OCEAN                             ENEMY WATERS
   A B C D E F G H I J                  A B C D E F G H I J
 1 C C C C C ~ ~ ~ ~ ~               1 ~ ~ ~ ~ ~ ~ ~ ~ ~ ~
 2 ~ ~ ~ ~ ~ ~ ~ ~ ~ ~               2 ~ ~ ~ ~ X ~ ~ ~ ~ ~
 3 ~ ~ ~ ~ ~ ~ ~ ~ ~ ~               3 ~ ~ ~ ~ ○ ~ ~ ~ ~ ~
...

Status: Ships remaining - You: 5 | Computer: 4
Your accuracy: 50.0% (4/8)
```

## Tests

Run focused tests for the Battleship module:

```bash
cd battleship
go test ./...
```
