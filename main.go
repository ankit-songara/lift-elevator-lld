package main // entry point of the program

import (
	"fmt"  // for printing output
	"math" // for math.Abs (absolute value) and math.MaxInt32
	"sync" // for sync.Mutex (concurrency lock)
)

// Direction is a custom type based on int (like an enum)
type Direction int

const (
	IDLE Direction = 0  // elevator not moving
	UP   Direction = 1  // moving up   → currentFloor += 1
	DOWN Direction = -1 // moving down → currentFloor += -1
)

// DoorState is a custom type for door open/close
type DoorState int

const (
	OPEN   DoorState = 0 // door is open
	CLOSED DoorState = 1 // door is closed
)

// Elevator represents one elevator in the building
type Elevator struct {
	mu            sync.Mutex   // lock — prevents two goroutines editing same elevator at once
	id            int          // elevator number (1, 2, 3...)
	currentFloor  int          // which floor elevator is currently on
	direction     Direction    // UP, DOWN, or IDLE
	doorState     DoorState    // OPEN or CLOSED
	floorRequests map[int]bool // set of floors to stop at. map[5]=true means "stop at floor 5"
}

// NewElevator creates a new elevator starting at startFloor
func NewElevator(id, startFloor int) *Elevator { // returns a pointer (*) to avoid copying
	return &Elevator{ // & gives address of the struct (pointer)
		id:            id,
		currentFloor:  startFloor,
		direction:     IDLE,               // starts idle
		doorState:     CLOSED,             // starts with door closed
		floorRequests: make(map[int]bool), // make() initializes the map (must do this or crash)
	}
}

// AddRequest adds a floor to this elevator's stop list
func (e *Elevator) AddRequest(floor int) { // (e *Elevator) = method on Elevator, like "this"
	e.mu.Lock()         // acquire lock — block other goroutines from entering
	defer e.mu.Unlock() // defer = run this when function exits (auto unlock, always safe)

	e.floorRequests[floor] = true // add floor to the request set

	if e.direction == IDLE { // if elevator not moving, set its direction now
		if floor > e.currentFloor {
			e.direction = UP // requested floor is above → go up
		} else if floor < e.currentFloor {
			e.direction = DOWN // requested floor is below → go down
		}
		// if floor == currentFloor, stay IDLE (door opens in Step)
	}

	fmt.Printf("Elevator %d: Request added for floor %d\n", e.id, floor) // %d = integer placeholder
}

// OpenDoor opens the elevator door
func (e *Elevator) OpenDoor() {
	e.mu.Lock() // lock before modifying state
	defer e.mu.Unlock()
	e.doorState = OPEN
	fmt.Printf("Elevator %d: Door OPEN at floor %d\n", e.id, e.currentFloor)
}

// CloseDoor closes the elevator door
func (e *Elevator) CloseDoor() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.doorState = CLOSED
	fmt.Printf("Elevator %d: Door CLOSED at floor %d\n", e.id, e.currentFloor)
}

// Step moves the elevator one unit — call this repeatedly to simulate movement
func (e *Elevator) Step() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.floorRequests) == 0 { // len() on map = number of keys = pending requests
		e.direction = IDLE // nothing to do → go idle
		return             // exit early
	}

	if e.floorRequests[e.currentFloor] { // if current floor is in request set
		delete(e.floorRequests, e.currentFloor) // remove it — request served
		fmt.Printf("Elevator %d: Stopped at floor %d\n", e.id, e.currentFloor)
		return // don't move this step, just stop and open door
	}

	// move elevator one floor in current direction
	// UP=1 so floor+1, DOWN=-1 so floor-1 — the math handles both cases
	e.currentFloor += int(e.direction) // int() converts Direction type to raw int
	fmt.Printf("Elevator %d: Moving to floor %d\n", e.id, e.currentFloor)

	// check if remaining requests are above or below current floor
	hasAbove, hasBelow := false, false   // := declares AND assigns (short form)
	for floor := range e.floorRequests { // range over map keys (floor numbers)
		if floor > e.currentFloor {
			hasAbove = true
		}
		if floor < e.currentFloor {
			hasBelow = true
		}
	}

	// SCAN algorithm: decide if we need to reverse direction
	switch { // switch without condition = cleaner if-else chain
	case e.direction == UP && !hasAbove && hasBelow:
		e.direction = DOWN // going up, nothing above, but something below → reverse
	case e.direction == DOWN && !hasBelow && hasAbove:
		e.direction = UP // going down, nothing below, but something above → reverse
	case !hasAbove && !hasBelow:
		e.direction = IDLE // nothing anywhere → stop
	}
}

// CurrentFloor safely returns the current floor (locks before reading)
func (e *Elevator) CurrentFloor() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.currentFloor // even reads need a lock — prevents race conditions
}

// Direction safely returns the current direction
func (e *Elevator) Direction() Direction {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.direction
}

// ElevatorController manages all elevators and assigns requests
type ElevatorController struct {
	mu        sync.Mutex  // lock for the controller itself
	elevators []*Elevator // slice (dynamic array) of elevator pointers
}

// NewElevatorController creates N elevators all starting at floor 1
func NewElevatorController(numElevators, numFloors int) *ElevatorController {
	elevators := make([]*Elevator, numElevators) // make slice of length numElevators, all nil initially
	for i := 0; i < numElevators; i++ {          // classic for loop (Go has no while)
		elevators[i] = NewElevator(i+1, 1) // elevator IDs start at 1, all start at floor 1
	}
	return &ElevatorController{elevators: elevators} // return pointer to controller
}

// score calculates how suitable an elevator is for a request
// lower score = better candidate
func (ec *ElevatorController) score(e *Elevator, floor int, dir Direction) int {
	curFloor := e.CurrentFloor() // safely get current floor
	curDir := e.Direction()      // safely get current direction

	// math.Abs needs float64, so convert → get absolute distance → convert back to int
	dist := int(math.Abs(float64(curFloor - floor)))

	if curDir == IDLE {
		return dist // idle elevator: cost is just the distance, no direction penalty
	}

	if curDir == UP && dir == UP && floor >= curFloor {
		return dist // elevator going UP, request is UP, floor is ahead → perfect, just distance
	}

	if curDir == DOWN && dir == DOWN && floor <= curFloor {
		return dist // elevator going DOWN, request is DOWN, floor is ahead → perfect, just distance
	}

	return 1000 + dist // wrong direction or past the floor → huge penalty, last resort
}

// HandleExternalRequest handles UP/DOWN button press on a floor
func (ec *ElevatorController) HandleExternalRequest(floor int, dir Direction) {
	ec.mu.Lock() // lock controller while we pick the best elevator
	defer ec.mu.Unlock()

	var best *Elevator         // var declares pointer, initialized to nil
	bestScore := math.MaxInt32 // start with worst possible score so any real score beats it

	for _, e := range ec.elevators { // _ discards index, e is each elevator
		if s := ec.score(e, floor, dir); s < bestScore { // s is scoped to this if block
			bestScore = s // found a better elevator
			best = e      // remember it
		}
	}

	if best != nil { // nil check — make sure we found an elevator
		best.AddRequest(floor) // assign the request to the best elevator
		fmt.Printf("Controller: Floor %d request assigned to Elevator %d\n", floor, best.id)
	}
}

// HandleInternalRequest handles floor button press inside an elevator
func (ec *ElevatorController) HandleInternalRequest(elevatorID, floor int) {
	for _, e := range ec.elevators { // search all elevators
		if e.id == elevatorID { // find the one with matching ID
			e.AddRequest(floor) // add the floor request to it
			return              // done, stop searching
		}
	}
}

// Step advances all elevators by one unit of movement
func (ec *ElevatorController) Step() {
	for _, e := range ec.elevators {
		e.Step() // each elevator moves independently
	}
}

func main() {
	controller := NewElevatorController(3, 10) // 3 elevators, 10 floor building

	controller.HandleExternalRequest(3, UP)   // someone on floor 3 pressed UP
	controller.HandleExternalRequest(7, DOWN) // someone on floor 7 pressed DOWN
	controller.HandleInternalRequest(1, 8)    // person inside elevator 1 pressed floor 8

	for i := 0; i < 10; i++ { // simulate 10 steps
		fmt.Printf("\n--- Step %d ---\n", i+1)
		controller.Step() // move all elevators one step
	}
}
