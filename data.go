package kaa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MetaData
// contains any computed data about the move request.
// is used in composition with the move request so you cannot
// have name colisions with the MoveRequest struct
type MetaData struct {
	// denotes the number of moves until you reach the closest piece of food
	MyLength   int
	Hazards    map[string]bool
	FoodMap    map[string]bool
	tightSpace bool
	// making this a pointer makes it able to be tested against
	// nil so we might as well keep it like this
	SnakeHash  map[string]*SnakeData
	SnakeHeads map[string]bool
}

// MetaDataDirec
// contains any computed data in a particular direction
// is used in composition with the move request so you cannot
// have name colisions with the MoveRequest struct
type MetaDataDirec struct {
	// denotes the number of moves until you reach the closest piece of food
	ClosestFood int
	// totals up your length and the ammount of food in a direction
	// if you would fill up the space make it unlikely to go that direction
	MovesVsSpace int
	// the total number of moves possible in this direction
	TotalMoves int
	TotalFood  int
	// contains a map to the last accessable piece of a snake
	// from your current location if you moved in this direction
	KeySnakeData map[int]*SnakeData
}

// minKeySnakePart
// returns the snake data for the point you are waiting to open up
// it is the least number of moves that anyone around you can make before
// you are able to exit the area you are in
func (m *MetaDataDirec) minKeySnakePart() *SnakeData {
	var min *SnakeData
	for _, val := range m.KeySnakeData {
		if min == nil || min.lengthLeft > val.lengthLeft {
			min = val
		}
	}
	return min
}

func (m *MoveRequest) NoFood() bool {
	for _, val := range m.Direcs {
		if val.TotalFood > 0 {
			return false
		}
	}
	return true
}

// String
// used to print the metadata for a particular direction
// it is necessary because the Static data is a pointer
// unfortunately this means that you have to manually manage this
// maybe I could make
func (m *MetaDataDirec) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("MetaDataDirec{")
	buffer.WriteString(fmt.Sprintf("ClosestFood:%v, ", m.ClosestFood))
	buffer.WriteString(fmt.Sprintf("movesVsSpace:%v, ", m.MovesVsSpace))
	buffer.WriteString(fmt.Sprintf("TotalMoves:%v, ", m.TotalMoves))
	buffer.WriteString(fmt.Sprintf("KeySnakeData:\n"))
	for direc, val := range m.KeySnakeData {
		buffer.WriteString(fmt.Sprintf("\t%v:%v", direc, val))
	}
	buffer.WriteString("}\n")

	return buffer.String()
}

// used to find and set the length of your snake globally in the
// metatdata object
func (m *MetaData) SetMyLength(data *MoveRequest) {
	for _, snake := range data.Snakes {
		if snake.Id == data.You && len(data.You) > 0 {
			m.MyLength = len(snake.Coords)
		}
	}
}

// a little struct used to see the length left after this portion of a
// snakes body the tail of the snake has a value of 1
type SnakeData struct {
	id         int
	lengthLeft int
	pnt        *Point
}

func (s *SnakeData) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	buffer.WriteString(fmt.Sprintf("id:%v, ", s.id))
	buffer.WriteString(fmt.Sprintf("lengthLeft:%v, ", s.lengthLeft))
	buffer.WriteString(fmt.Sprintf("pnt:%v, ", s.pnt))
	buffer.WriteString("}\n")

	return buffer.String()
}

// GenenSnakeHash
//	generates a map of all the points in all the snakes
//	is used to determine how much of the snake must move
//      in order for the area they are blocking to be open
func (m *MetaData) GenSnakeHash(data *MoveRequest) {
	m.SnakeHash = make(map[string]*SnakeData)
	for i, snake := range data.Snakes {
		for j, coord := range snake.Coords {
			m.SnakeHash[coord.String()] = &SnakeData{
				id:         i,
				lengthLeft: len(snake.Coords) - j - 1,
				pnt:        &Point{coord.X, coord.Y},
			}
		}
	}
}

// Generates a map of hazards
func (m *MetaData) GenHazards(data *MoveRequest) {
	m.Hazards = make(map[string]bool)
	for _, snake := range data.Snakes {
		if len(snake.Coords) >= m.MyLength && data.You != snake.Id {
			head := snake.Head()
			d := head.Down(data)
			if d != nil {
				m.Hazards[d.String()] = true
			}
			d = head.Up(data)
			if d != nil {
				m.Hazards[d.String()] = true
			}
			d = head.Right(data)
			if d != nil {
				m.Hazards[d.String()] = true
			}
			d = head.Left(data)
			if d != nil {
				m.Hazards[d.String()] = true
			}

		}
		for _, coord := range snake.Coords {
			m.Hazards[coord.String()] = true
		}
	}
}

// generates a map of all the food points
func (m *MetaData) GenFoodMap(data *MoveRequest) {
	m.FoodMap = make(map[string]bool)
	for _, food := range data.Food {
		m.FoodMap[food.String()] = true
	}
}

// alias for the metadata map
type MoveMetaData map[string]*MetaDataDirec

// StaticData
// a list of found information in a direction is used in a breadth
// first search to determine the ammount of food you can reach in
// a desired number of moves from the source
type StaticData struct {
	ClosestFood *Point
	Food        int
	Snakes      int
	Moves       int
}

func (m *StaticData) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\nStaticData{")
	buffer.WriteString(fmt.Sprintf("ClosestFood:%v, ", m.ClosestFood))
	buffer.WriteString(fmt.Sprintf("Food:%v, ", m.Food))
	buffer.WriteString(fmt.Sprintf("Snakes:%v, ", m.Snakes))
	buffer.WriteString(fmt.Sprintf("Moves:%v, ", m.Moves))
	buffer.WriteString("}\n")

	return buffer.String()
}

// RESPONSE AND REQUEST STRUCTS
type MoveResponse struct {
	Move  string `json:"move"`
	Taunt string `json:"taunt,omitempty"`
}

type GameStartRequest struct {
	GameId string `json:"game_id"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

func NewGameStartRequest(req *http.Request) (*GameStartRequest, error) {
	decoded := GameStartRequest{}
	err := json.NewDecoder(req.Body).Decode(&decoded)
	return &decoded, err
}

type GameStartResponse struct {
	Color   string  `json:"color"`
	HeadUrl *string `json:"head_url,omitempty"`
	Name    string  `json:"name"`
	Taunt   *string `json:"taunt,omitempty"`
}

type MoveRequest struct {
	// static
	GameId string  `json:"game_id"`
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Turn   int     `json:"turn"`
	Food   []Point `json:"food"`
	Snakes []Snake `json:"snakes"`
	You    string  `json:"you"`

	// added here for convenience
	MetaData
	Direcs MoveMetaData
}

func (m *MoveRequest) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("\nMoveRequest{\n")
	head, _ := getMyHead(m)
	buffer.WriteString(fmt.Sprintf("head: %v ", head))
	buffer.WriteString("Direcs{\n")
	for direc, val := range m.Direcs {
		buffer.WriteString(fmt.Sprintf("\t%v:%v", direc, val))
	}
	buffer.WriteString(fmt.Sprintf("tightSpace: %v ", m.tightSpace))
	buffer.WriteString("}\n")
	buffer.WriteString("}\n")

	return buffer.String()
}

// initializes global meta data attributes
func (m *MoveRequest) init() {
	m.SetMyLength(m)
	m.GenHazards(m)
	m.GenFoodMap(m)
	m.GenSnakeHash(m)
}

// de serializes the move request data into a string
func getMoveRequestString(req *http.Request) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	return buf.String()
}

// creates a new move request
func NewMoveRequest(str string) (*MoveRequest, error) {
	res := new(MoveRequest)
	err := json.Unmarshal([]byte(str), res)
	if err != nil {
		return nil, err
	}
	err = GenerateMetaData(res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Decode [number, number] JSON array into a Point

// Allows decoding a string or number identifier in JSON
// by removing any surrounding quotes and storing in a string
type Snake struct {
	Coords       []Point `json:"coords"`
	HealthPoints int     `json:"health_points"`
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Taunt        string  `json:"taunt"`
}

func (snake Snake) Head() *Point    { return &snake.Coords[0] }
func (snake *Snake) String() string { return fmt.Sprintf("%#v", snake) }
