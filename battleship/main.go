package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const boardSize = 10
const (
	waterGlyph = "~"
	hitGlyph   = "X"
	missGlyph  = "○"
)

type Coord struct {
	Row int
	Col int
}

type ShotOutcome int

const (
	Unknown ShotOutcome = iota
	Miss
	Hit
)

type ShipSpec struct {
	Name   string
	Size   int
	Symbol rune
	Art    []string
}

type Ship struct {
	Name   string
	Size   int
	Symbol rune
	Cells  []Coord
	Hits   int
	Art    []string
}

type Cell struct {
	Ship *Ship
	Hit  bool
}

type Board struct {
	Cells [boardSize][boardSize]Cell
	Ships []*Ship
}

type Stats struct {
	Shots int
	Hits  int
}

type AI struct {
	rand      *rand.Rand
	available []Coord
	shotSet   map[Coord]bool
	targets   []Coord
	hitStreak []Coord
}

var fleet = []ShipSpec{
	{Name: "Carrier", Size: 5, Symbol: 'C', Art: []string{"   __/___", " _____/______|", " \\", " ~~~~~~~~~~~~~~~"}},
	{Name: "Battleship", Size: 4, Symbol: 'B', Art: []string{"   |\\", " __|_\\__", "|  BATTLE |", "~~~~~~~~~~~"}},
	{Name: "Cruiser", Size: 3, Symbol: 'R', Art: []string{"    /\\", " __/==\\__", "|  CRUISE |", "~~~~~~~~~~~"}},
	{Name: "Submarine", Size: 3, Symbol: 'S', Art: []string{"   ___", " _/___\\_", "|_SUB___|", "  \\___/"}},
	{Name: "Destroyer", Size: 2, Symbol: 'D', Art: []string{"   /_\\", " _|D |__", "~~~~~~~~"}},
}

const (
	blue   = "\033[34m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	gray   = "\033[37m"
	bold   = "\033[1m"
	reset  = "\033[0m"
)

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	reader := bufio.NewReader(os.Stdin)
	_, noColor := os.LookupEnv("NO_COLOR")
	color := !noColor

	showTitle(color)
	promptEnter(reader, "Press Enter to start your naval campaign...")

	playerBoard := &Board{}
	computerBoard := &Board{}
	ai := NewAI(r)

	placePlayerShips(reader, playerBoard, color)
	placeComputerShips(computerBoard, r)

	var playerTracking [boardSize][boardSize]ShotOutcome
	playerStats := Stats{}
	computerStats := Stats{}

	for {
		clearScreen()
		showBoards(playerBoard, playerTracking, color)
		showStatus(playerBoard, computerBoard, playerStats, computerStats, color)

		shot := promptShot(reader, "Fire at enemy (e.g. A5, or 'help'):", playerTracking)
		if shot.Row == -1 {
			showHelp(color)
			promptEnter(reader, "Press Enter to continue...")
			continue
		}
		hit, sunk, _ := computerBoard.ReceiveShot(shot)
		playerStats.Shots++
		if hit {
			playerStats.Hits++
			playerTracking[shot.Row][shot.Col] = Hit
			animateShot("Direct hit!", color, true)
			if sunk != nil {
				announceSunk(*sunk, true, color)
			}
		} else {
			playerTracking[shot.Row][shot.Col] = Miss
			animateShot("Splash... miss.", color, false)
		}
		if computerBoard.AllSunk() {
			clearScreen()
			showBoards(playerBoard, playerTracking, color)
			showEndGame(true, playerStats, computerStats, color)
			return
		}

		compShot := ai.NextShot()
		fmt.Printf("\nComputer fires at %s...\n", coordToString(compShot))
		time.Sleep(450 * time.Millisecond)
		hit, sunk, _ = playerBoard.ReceiveShot(compShot)
		computerStats.Shots++
		if hit {
			computerStats.Hits++
			fmt.Println(colorize(red, "Enemy hit your ship!"))
			if sunk != nil {
				announceSunk(*sunk, false, color)
			}
		} else {
			fmt.Println(colorize(gray, "Enemy missed."))
		}
		ai.RegisterResult(compShot, hit, sunk != nil)

		if playerBoard.AllSunk() {
			clearScreen()
			showBoards(playerBoard, playerTracking, color)
			showEndGame(false, playerStats, computerStats, color)
			return
		}
		promptEnter(reader, "Press Enter for next turn...")
	}
}

func NewAI(r *rand.Rand) *AI {
	coords := make([]Coord, 0, boardSize*boardSize)
	for i := 0; i < boardSize; i++ {
		for j := 0; j < boardSize; j++ {
			coords = append(coords, Coord{Row: i, Col: j})
		}
	}
	r.Shuffle(len(coords), func(i, j int) { coords[i], coords[j] = coords[j], coords[i] })
	return &AI{
		rand:      r,
		available: coords,
		shotSet:   map[Coord]bool{},
	}
}

func (ai *AI) NextShot() Coord {
	for len(ai.targets) > 0 {
		n := ai.targets[0]
		ai.targets = ai.targets[1:]
		if !ai.shotSet[n] {
			ai.shotSet[n] = true
			return n
		}
	}
	for len(ai.available) > 0 {
		n := ai.available[len(ai.available)-1]
		ai.available = ai.available[:len(ai.available)-1]
		if !ai.shotSet[n] {
			ai.shotSet[n] = true
			return n
		}
	}
	return Coord{}
}

func (ai *AI) RegisterResult(c Coord, hit bool, sunk bool) {
	if !hit {
		return
	}
	if sunk {
		ai.targets = nil
		ai.hitStreak = nil
		return
	}
	ai.hitStreak = append(ai.hitStreak, c)
	ai.rebuildTargets()
}

func (ai *AI) rebuildTargets() {
	candidates := make([]Coord, 0)
	if len(ai.hitStreak) >= 2 {
		alignedRow := true
		alignedCol := true
		row := ai.hitStreak[0].Row
		col := ai.hitStreak[0].Col
		minR, maxR := row, row
		minC, maxC := col, col
		for _, h := range ai.hitStreak[1:] {
			if h.Row != row {
				alignedRow = false
			}
			if h.Col != col {
				alignedCol = false
			}
			if h.Row < minR {
				minR = h.Row
			}
			if h.Row > maxR {
				maxR = h.Row
			}
			if h.Col < minC {
				minC = h.Col
			}
			if h.Col > maxC {
				maxC = h.Col
			}
		}
		if alignedRow {
			candidates = append(candidates, Coord{Row: row, Col: minC - 1}, Coord{Row: row, Col: maxC + 1})
		} else if alignedCol {
			candidates = append(candidates, Coord{Row: minR - 1, Col: col}, Coord{Row: maxR + 1, Col: col})
		}
	}
	if len(candidates) == 0 {
		for _, h := range ai.hitStreak {
			candidates = append(candidates,
				Coord{Row: h.Row - 1, Col: h.Col},
				Coord{Row: h.Row + 1, Col: h.Col},
				Coord{Row: h.Row, Col: h.Col - 1},
				Coord{Row: h.Row, Col: h.Col + 1},
			)
		}
	}
	seen := map[Coord]bool{}
	next := make([]Coord, 0)
	for _, c := range candidates {
		if c.Row < 0 || c.Row >= boardSize || c.Col < 0 || c.Col >= boardSize {
			continue
		}
		if ai.shotSet[c] || seen[c] {
			continue
		}
		seen[c] = true
		next = append(next, c)
	}
	ai.rand.Shuffle(len(next), func(i, j int) { next[i], next[j] = next[j], next[i] })
	ai.targets = next
}

func (b *Board) CanPlace(start Coord, horizontal bool, size int) bool {
	for i := 0; i < size; i++ {
		r, c := start.Row, start.Col
		if horizontal {
			c += i
		} else {
			r += i
		}
		if r < 0 || r >= boardSize || c < 0 || c >= boardSize {
			return false
		}
		if b.Cells[r][c].Ship != nil {
			return false
		}
	}
	return true
}

func (b *Board) PlaceShip(spec ShipSpec, start Coord, horizontal bool) bool {
	if !b.CanPlace(start, horizontal, spec.Size) {
		return false
	}
	ship := &Ship{Name: spec.Name, Size: spec.Size, Symbol: spec.Symbol, Art: spec.Art}
	for i := 0; i < spec.Size; i++ {
		r, c := start.Row, start.Col
		if horizontal {
			c += i
		} else {
			r += i
		}
		b.Cells[r][c].Ship = ship
		ship.Cells = append(ship.Cells, Coord{Row: r, Col: c})
	}
	b.Ships = append(b.Ships, ship)
	return true
}

func (b *Board) ReceiveShot(c Coord) (hit bool, sunk *Ship, already bool) {
	cell := &b.Cells[c.Row][c.Col]
	if cell.Hit {
		return false, nil, true
	}
	cell.Hit = true
	if cell.Ship == nil {
		return false, nil, false
	}
	cell.Ship.Hits++
	if cell.Ship.Hits == cell.Ship.Size {
		return true, cell.Ship, false
	}
	return true, nil, false
}

func (b *Board) AllSunk() bool {
	for _, s := range b.Ships {
		if s.Hits < s.Size {
			return false
		}
	}
	return true
}

func parseCoordinate(input string) (Coord, error) {
	in := strings.ToUpper(strings.TrimSpace(input))
	if len(in) < 2 || len(in) > 3 {
		return Coord{}, fmt.Errorf("invalid coordinate format (use A1-J10)")
	}
	col := int(in[0] - 'A')
	if col < 0 || col >= boardSize {
		return Coord{}, fmt.Errorf("column must be A-J")
	}
	rowN, err := strconv.Atoi(in[1:])
	if err != nil || rowN < 1 || rowN > boardSize {
		return Coord{}, fmt.Errorf("row must be 1-10")
	}
	return Coord{Row: rowN - 1, Col: col}, nil
}

func coordToString(c Coord) string {
	return fmt.Sprintf("%c%d", rune('A'+c.Col), c.Row+1)
}

func placePlayerShips(reader *bufio.Reader, board *Board, color bool) {
	for _, spec := range fleet {
		for {
			clearScreen()
			showTitle(color)
			fmt.Println(colorize(yellow, "Place your fleet:"))
			printBoardLines(board.RenderOwn(color))
			fmt.Printf("\nDeploy %s (size %d)\n", spec.Name, spec.Size)
			printArt(spec.Art, colorize(green, ""), color)
			coord := promptText(reader, "Start coordinate (A1-J10):")
			c, err := parseCoordinate(coord)
			if err != nil {
				fmt.Println(colorize(yellow, err.Error()))
				promptEnter(reader, "Press Enter to retry...")
				continue
			}
			o := strings.ToUpper(promptText(reader, "Orientation [H/V]:"))
			horizontal := o == "H"
			if o != "H" && o != "V" {
				fmt.Println(colorize(yellow, "Orientation must be H or V"))
				promptEnter(reader, "Press Enter to retry...")
				continue
			}
			if !board.PlaceShip(spec, c, horizontal) {
				fmt.Println(colorize(yellow, "Invalid placement: out-of-bounds or overlapping ship."))
				promptEnter(reader, "Press Enter to retry...")
				continue
			}
			break
		}
	}
}

func placeComputerShips(board *Board, r *rand.Rand) {
	for _, spec := range fleet {
		for {
			start := Coord{Row: r.Intn(boardSize), Col: r.Intn(boardSize)}
			horizontal := r.Intn(2) == 0
			if board.PlaceShip(spec, start, horizontal) {
				break
			}
		}
	}
}

func promptShot(reader *bufio.Reader, prompt string, tracking [boardSize][boardSize]ShotOutcome) Coord {
	for {
		in := strings.TrimSpace(promptText(reader, prompt))
		if strings.EqualFold(in, "help") {
			return Coord{Row: -1, Col: -1}
		}
		c, err := parseCoordinate(in)
		if err != nil {
			fmt.Println(colorize(yellow, err.Error()))
			continue
		}
		if tracking[c.Row][c.Col] != Unknown {
			fmt.Println(colorize(yellow, "You already fired at that coordinate."))
			continue
		}
		return c
	}
}

func showBoards(playerBoard *Board, tracking [boardSize][boardSize]ShotOutcome, color bool) {
	left := playerBoard.RenderOwn(color)
	right := renderTracking(tracking, color)
	fmt.Println(colorize(bold, "YOUR OCEAN") + strings.Repeat(" ", 28) + colorize(bold, "ENEMY WATERS"))
	for i := 0; i < len(left) && i < len(right); i++ {
		fmt.Printf("%-38s   %s\n", left[i], right[i])
	}
}

func showStatus(playerBoard, computerBoard *Board, p, c Stats, color bool) {
	playerRemain := shipsRemaining(playerBoard)
	computerRemain := shipsRemaining(computerBoard)
	fmt.Printf("\n%s Ships remaining - You: %d | Computer: %d\n", colorize(yellow, "Status:"), playerRemain, computerRemain)
	if p.Shots > 0 {
		fmt.Printf("Your accuracy: %.1f%% (%d/%d)\n", hitPercent(p), p.Hits, p.Shots)
	}
	if c.Shots > 0 {
		fmt.Printf("Computer accuracy: %.1f%% (%d/%d)\n", hitPercent(c), c.Hits, c.Shots)
	}
}

func shipsRemaining(b *Board) int {
	rem := 0
	for _, s := range b.Ships {
		if s.Hits < s.Size {
			rem++
		}
	}
	return rem
}

func (b *Board) RenderOwn(color bool) []string {
	lines := []string{headerLine()}
	for r := 0; r < boardSize; r++ {
		line := fmt.Sprintf("%2d ", r+1)
		for c := 0; c < boardSize; c++ {
			cell := b.Cells[r][c]
			glyph := waterGlyph
			style := blue
			if cell.Hit && cell.Ship != nil {
				glyph = hitGlyph
				style = red
			} else if cell.Hit {
				glyph = missGlyph
				style = gray
			} else if cell.Ship != nil {
				glyph = string(cell.Ship.Symbol)
				style = green
			}
			line += colorizeMaybe(style, glyph, color) + " "
		}
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return lines
}

func renderTracking(tracking [boardSize][boardSize]ShotOutcome, color bool) []string {
	lines := []string{headerLine()}
	for r := 0; r < boardSize; r++ {
		line := fmt.Sprintf("%2d ", r+1)
		for c := 0; c < boardSize; c++ {
			glyph := waterGlyph
			style := blue
			switch tracking[r][c] {
			case Hit:
				glyph = hitGlyph
				style = red
			case Miss:
				glyph = missGlyph
				style = gray
			}
			line += colorizeMaybe(style, glyph, color) + " "
		}
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return lines
}

func headerLine() string {
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	return "   " + strings.Join(cols, " ")
}

func promptText(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt + " ")
	in, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, "failed to read input from terminal:", err)
	}
	return strings.TrimSpace(in)
}

func promptEnter(reader *bufio.Reader, msg string) {
	fmt.Print(msg)
	_, _ = reader.ReadString('\n')
}

func showTitle(color bool) {
	fmt.Println(colorizeMaybe(blue, `
██████╗  █████╗ ████████╗████████╗██╗     ███████╗███████╗██╗  ██╗██╗██████╗
██╔══██╗██╔══██╗╚══██╔══╝╚══██╔══╝██║     ██╔════╝██╔════╝██║  ██║██║██╔══██╗
██████╔╝███████║   ██║      ██║   ██║     █████╗  ███████╗███████║██║██████╔╝
██╔══██╗██╔══██║   ██║      ██║   ██║     ██╔══╝  ╚════██║██╔══██║██║██╔═══╝
██████╔╝██║  ██║   ██║      ██║   ███████╗███████╗███████║██║  ██║██║██║
╚═════╝ ╚═╝  ╚═╝   ╚═╝      ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝╚═╝╚═╝
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~`, color))
}

func showHelp(color bool) {
	fmt.Println(colorizeMaybe(yellow, "\nCOMMAND HELP", color))
	fmt.Println("- Enter coordinates like A5 or J10 to fire.")
	fmt.Println("- Enter 'help' anytime during your turn to view this panel.")
	fmt.Println("- During placement, choose H (horizontal) or V (vertical).")
	fmt.Println("- Goal: sink all 5 enemy ships before yours are sunk.")
}

func animateShot(msg string, color bool, hit bool) {
	frames := []string{".", "..", "..."}
	for _, f := range frames {
		fmt.Printf("%s%s\r", msg, f)
		time.Sleep(130 * time.Millisecond)
	}
	if hit {
		fmt.Println(colorizeMaybe(red, msg+" 💥", color))
	} else {
		fmt.Println(colorizeMaybe(gray, msg+" ○", color))
	}
}

func announceSunk(ship Ship, playerDidIt bool, color bool) {
	if playerDidIt {
		fmt.Println(colorizeMaybe(yellow, "\nYou sank the enemy "+ship.Name+"!", color))
	} else {
		fmt.Println(colorizeMaybe(yellow, "\nYour "+ship.Name+" has been sunk!", color))
	}
	printArt(ship.Art, "", color)
}

func printArt(art []string, prefix string, color bool) {
	for _, line := range art {
		fmt.Println(prefix + colorizeMaybe(yellow, line, color))
	}
}

func showEndGame(playerWon bool, p, c Stats, color bool) {
	if playerWon {
		fmt.Println(colorizeMaybe(green, "\nVICTORY! All enemy ships destroyed.", color))
	} else {
		fmt.Println(colorizeMaybe(red, "\nDEFEAT! Your fleet has been sunk.", color))
	}
	fmt.Println("\nBattle Summary")
	fmt.Printf("You:      shots=%d hits=%d accuracy=%.1f%%\n", p.Shots, p.Hits, hitPercent(p))
	fmt.Printf("Computer: shots=%d hits=%d accuracy=%.1f%%\n", c.Shots, c.Hits, hitPercent(c))
}

func hitPercent(s Stats) float64 {
	if s.Shots == 0 {
		return 0
	}
	return float64(s.Hits) * 100 / float64(s.Shots)
}

func colorize(colorCode, text string) string {
	return colorCode + text + reset
}

func colorizeMaybe(colorCode, text string, enabled bool) string {
	if !enabled {
		return text
	}
	return colorCode + text + reset
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func printBoardLines(lines []string) {
	for _, line := range lines {
		fmt.Println(line)
	}
}
