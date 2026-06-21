# Lift Elevator LLD (Low-Level Design)

A complete low-level design implementation of an elevator system in Go, demonstrating concurrent programming with goroutines, thread-safety with mutexes, and the SCAN algorithm for elevator scheduling.

## Overview

This project implements a realistic elevator control system for a multi-floor building with multiple elevators. It showcases:

- **Concurrency**: Multiple elevators moving simultaneously using goroutines
- **Thread-Safety**: Mutex locks to prevent race conditions
- **Algorithm**: SCAN algorithm for efficient elevator routing
- **State Management**: Elevator states (IDLE, UP, DOWN) and door states (OPEN, CLOSED)

## Key Features

### 1. **Elevator State Management**
- **Direction**: UP, DOWN, or IDLE
- **Door State**: OPEN or CLOSED
- **Current Floor**: Tracks position in the building
- **Floor Requests**: Queue of floors to visit

### 2. **SCAN Algorithm**
The elevator uses the SCAN algorithm for optimal routing:
- Goes UP until no more floors above, then reverses to go DOWN
- Minimizes wait time and distance traveled
- Prevents "starvation" (some floors never getting service)

### 3. **Thread-Safe Operations**
- `sync.Mutex` locks protect elevator state
- Prevents concurrent modification issues
- Safe for multiple goroutines accessing the same elevator

### 4. **Smart Request Handling**
- **External Requests**: UP/DOWN button press on a floor
- **Internal Requests**: Floor button press inside the elevator
- **Scoring System**: Picks the best elevator for each request

## Project Structure

```
Lift Elevator LLD/
├── main.go            # Complete implementation
├── go.mod             # Module definition
└── README.md          # This file
```

## Code Architecture

### Core Types

#### `Direction`
```go
const (
    IDLE Direction = 0   // Not moving
    UP   Direction = 1   // Moving up
    DOWN Direction = -1  // Moving down
)
```

#### `Elevator`
- Manages individual elevator state and movements
- Handles floor requests and door operations
- Implements the SCAN algorithm in `Step()`

#### `ElevatorController`
- Manages multiple elevators
- Assigns requests to optimal elevator using scoring
- Orchestrates movement across all elevators

### Key Methods

| Method | Purpose |
|--------|---------|
| `AddRequest(floor)` | Add a floor to the elevator's stop list |
| `Step()` | Move elevator one unit; execute SCAN algorithm |
| `OpenDoor()` / `CloseDoor()` | Control door state |
| `HandleExternalRequest(floor, dir)` | Process UP/DOWN button on floor |
| `HandleInternalRequest(elevatorID, floor)` | Process button inside elevator |

## Running the Project

### Prerequisites

- Go 1.16 or higher

### Installation

```bash
git clone https://github.com/ankit-songara/lift-elevator-lld.git
cd "lift-elevator-lld"
```

### Build

```bash
go build -o lift-elevator-lld.exe main.go
```

### Run

```bash
go run main.go
```

or 

```bash
./lift-elevator-lld.exe
```

### Example Output

```
✓ Requested floor 3 (UP) → Assigned to Elevator 1
✓ Requested floor 7 (DOWN) → Assigned to Elevator 2
✓ Requested floor 8 (Internal to Elevator 1)

--- Step 1 ---
Elevator 1: Request added for floor 3
Elevator 1: Moving to floor 2
...

--- Step 10 ---
Elevator 1: Stopped at floor 8
```

## Design Patterns

### 1. **Factory Pattern**
- `NewElevator()`: Creates elevator instances
- `NewElevatorController()`: Creates the controller

### 2. **Strategy Pattern** (Implicit)
- Scoring algorithm for request assignment
- Different strategies could be swapped for elevator selection

### 3. **Locking/Concurrency Pattern**
- Mutex-based synchronization
- Defer for guaranteed unlock
- Pattern: `lock → defer unlock → modify state`

## Algorithms Explained

### SCAN Algorithm

The SCAN algorithm works like an elevator in a real building:

```
Initial State: Floor 5, direction UP
Requests: [3, 8, 2]

Step 1: Check if current floor requested
        → Not yet
        
Step 2: Move in current direction
        → Floor 6
        
Step 3: Check remaining requests
        - Above current: [8] ✓
        - Below current: [2, 3] ✓
        
Step 4: Continue UP (already has requests above)
        → Floor 7, Floor 8 (STOP)

Step 5: All requests above done, has below
        → Reverse direction to DOWN
        
Step 6: Continue DOWN
        → Floor 5, Floor 3 (STOP), Floor 2 (STOP)

Done: All requests served
```

### Scoring Algorithm

When a floor button is pressed, the elevator picks the best elevator:

```go
score = distance
if in_wrong_direction:
    score += 1000  // Heavy penalty
```

Lower score wins. An IDLE elevator with minimal distance is always chosen first.

## Thread Safety Example

```go
// Race condition WITHOUT mutex:
elevator.currentFloor = 5      // Goroutine A
floor := elevator.currentFloor // Goroutine B reads mid-write
// Could read corrupted value!

// Safe WITH mutex:
e.mu.Lock()
e.currentFloor = 5
e.mu.Unlock()

e.mu.Lock()
floor := e.currentFloor
e.mu.Unlock()
```

## Learning Outcomes

After studying this project, you'll understand:

- ✅ How real elevator systems work algorithmically
- ✅ Go's concurrency primitives (goroutines, channels, mutexes)
- ✅ Race condition prevention techniques
- ✅ State machine design in distributed systems
- ✅ Low-level system design principles
- ✅ Design patterns (Factory, Locking)

## Testing

The project includes a built-in simulation in `main()`. To extend:

1. Create additional request patterns
2. Add performance metrics (avg wait time, total distance)
3. Compare different algorithms
4. Test with high concurrency (100+ concurrent requests)

## Future Enhancements

- [ ] Add priority levels (emergency, express)
- [ ] Implement weight-based capacity constraints
- [ ] Add floor-specific rules (fire floor blocked)
- [ ] Performance metrics and visualization
- [ ] REST API for external control
- [ ] Log analysis and reporting

## Common Interview Questions

**Q: Why use the SCAN algorithm?**
A: It provides fairness (no starvation), minimizes avg wait time, and is simple to implement.

**Q: How would you handle an emergency?**
A: Add priority levels; emergency requests jump to front of queue, elevator moves immediately to that floor.

**Q: What if an elevator breaks down?**
A: Remove it from the `elevators` slice and rebalance requests to remaining elevators using the scoring system.

**Q: How to distribute loads between elevators?**
A: The scoring system already does this by picking the closest available elevator.

## Resources

- [SCAN Algorithm Explained](https://en.wikipedia.org/wiki/Elevator_algorithm)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Sync Package Documentation](https://pkg.go.dev/sync)

## License

MIT License

## Author

Ankit Songara

---

**Note**: This is an educational implementation. Production elevator systems use more sophisticated algorithms, redundancy, and safety mechanisms.
