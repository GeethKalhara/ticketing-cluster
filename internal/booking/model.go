package booking

import "fmt"

const (
	HoldTTLL        = 30000
	maxSeatsPerHold = 4
)

type SeatStatus string

const (
	Available SeatStatus = "AVAILABLE" // free to hold
	Held      SeatStatus = "HELD"      // reserved for one customer, for a short time
	Booked    SeatStatus = "BOOKED"    // sold
)

type Show struct {
	ID        string `json:"id"`
	TheaterID string `json:"theater_id"`
	MovieID   string `json:"movie_id"`
	Rows      int    `json:"rows"` // 10 -> rows A..J
	Cols      int    `json:"cols"` // 5  -> columns 1..5
}

type Seat struct {
	ShowID    string
	SeatID    string // "A1" .. "J5"
	Status    SeatStatus
	HeldBy    string // customer ID, only meaningful while Held
	HoldUntil int64  // when the hold ends (leader time, ms). Over at or after this.
	BookingID string // set once the seat is Booked
}

// Booking is the ticket a customer keeps after confirming.
type Booking struct {
	ID         string   `json:"id"`
	ShowID     string   `json:"show_id"`
	SeatIDs    []string `json:"seat_ids"` // a LIST: people book several seats together
	CustomerID string   `json:"customer_id"`
	LogIndex   uint64   `json:"log_index"` // position in the replicated log; proves the order
}

// ---------------------------------------------------------------------------
// Commands: the request forms
// ---------------------------------------------------------------------------

// CommandType says what the customer wants to do.
type CommandType string

const (
	CmdHold    CommandType = "HOLD"    // reserve seats for a short time
	CmdConfirm CommandType = "CONFIRM" // turn a hold into a booking
	CmdRelease CommandType = "RELEASE" // give a hold back
)

// Command is the request form that goes into the replicated log.
//
// RULE: everything the state machine needs to decide (including "what time is
// it") must be written INSIDE the command. Applying a command may not read a
// clock, use random numbers, or touch the network, because all 5 nodes must
// reach the same answer from the same form.
type Command struct {
	RequestID  string      `json:"request_id"` // unique per request; makes retries safe
	Type       CommandType `json:"type"`
	ShowID     string      `json:"show_id"`
	SeatIDs    []string    `json:"seat_ids"`
	CustomerID string      `json:"customer_id"`
	LeaderTime int64       `json:"leader_time"` // ms, stamped by the leader (Role 1)
	Lamport    uint64      `json:"lamport"`     // logical timestamp (Role 1); just carried here
}

// ---------------------------------------------------------------------------
// Results: the clerk's answer
// ---------------------------------------------------------------------------

// ResultCode is the outcome of applying one command.
type ResultCode string

const (
	OK           ResultCode = "OK"
	SeatTaken    ResultCode = "SEAT_TAKEN"     // someone else holds or owns a requested seat
	HoldExpired  ResultCode = "HOLD_EXPIRED"   // the hold ended before the confirm arrived
	NotHolder    ResultCode = "NOT_HOLDER"     // you are not the customer holding these seats
	NoSuchShow   ResultCode = "NO_SUCH_SHOW"   // unknown show ID
	NoSuchSeat   ResultCode = "NO_SUCH_SEAT"   // unknown seat ID in that show
	TooManySeats ResultCode = "TOO_MANY_SEATS" // more than MaxSeatsPerHold, or none at all
	BadRequest   ResultCode = "BAD_REQUEST"    // malformed command (e.g. no request ID)
)

// Result is the answer to one command.
//
// A refusal such as SEAT_TAKEN is a NORMAL result, not a Go error. The system
// worked correctly and the answer is "no". It is also stored under the request
// ID, so a retry of the same request gets the same answer.
type Result struct {
	Code      ResultCode `json:"code"`
	BookingID string     `json:"booking_id,omitempty"` // set when a CONFIRM succeeds
}

// SeatView is what a client sees when it asks for a show's seat map.
type SeatView struct {
	SeatID string     `json:"seat_id"`
	Status SeatStatus `json:"status"`
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

// SeatName turns zero-based row and column numbers into a seat name:
// (0,0) -> "A1", (1,2) -> "B3", (9,4) -> "J5".
func SeatName(row, col int) string {
	return fmt.Sprintf("%c%d", 'A'+row, col+1)
}

// seatKey builds the map key for a seat: "S1/A1". Lowercase first letter means
// it is private to this package.
func seatKey(showID, seatID string) string {
	return showID + "/" + seatID
}
