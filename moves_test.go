package dicebot

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
)

var botWithMoves = &Bot{
	db: &JsonDatabase{},
	moves: map[string]Move{
		"move": {
			Name:        "Move",
			Description: "When you do a thing, roll+Str.",
			Roll:        "2d6+Str",
			Hit:         "You successfully did the thing!",
			Pass:        "You did the thing, but something bad happens.",
			Miss:        "The thing went terribly wrong.",
		},
	},
}

func TestLoadMoves(t *testing.T) {
	json := `[
  {
    "name": "Move",
    "description": "When you do a thing, roll+Str.",
    "roll": "2d6+Str",
    "hit": "You successfully did the thing!",
    "pass": "You did the thing, but something bad happens.",
    "miss": "The thing went terribly wrong."
  }
]
`

	filename := WriteTempFile(t, "test*.json", json)
	defer os.Remove(filename)

	moves := make(map[string]Move)
	err := LoadMoves(moves, filename)
	if err != nil {
		t.Fatalf("LoadMoves(%v): %v", filename, err)
	}

	if !reflect.DeepEqual(moves, botWithMoves.moves) {
		t.Errorf("NewJsonDatabase(): expected %+v got %+v", botWithMoves.moves, moves)
	}
}

func ExampleBot_MakeMove_unknown() {
	fmt.Println(botWithMoves.MakeMove(context, "Unknown"))
	// Output:
	// Sorry, I don't understand how to parse 'Unknown': unknown move
}

func ExampleBot_MakeMove_error() {
	fmt.Println(botWithMoves.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll+Str.
	// Sorry, I don't understand how to parse '2d6+Str': undefined variable `Str`
}

func ExampleBot_MakeMove_hit() {
	rand.Seed(1)

	err := botWithMoves.db.StoreValue("str", "user-user", "1")
	if err != nil {
		log.Fatalf("StoreValue(str): %v", err)
	}

	fmt.Println(botWithMoves.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll+Str.
	// 2d6+Str => **(6 + 4) + 1** => **11**
	// You successfully did the thing!
}

func ExampleBot_MakeMove_pass() {
	rand.Seed(1)

	err := botWithMoves.db.StoreValue("str", "user-user", "-1")
	if err != nil {
		log.Fatalf("StoreValue(str): %v", err)
	}

	fmt.Println(botWithMoves.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll+Str.
	// 2d6+Str => **(6 + 4) + -1** => **9**
	// You did the thing, but something bad happens.
}

func ExampleBot_MakeMove_miss() {
	rand.Seed(1)

	err := botWithMoves.db.StoreValue("str", "user-user", "-4")
	if err != nil {
		log.Fatalf("StoreValue(str): %v", err)
	}

	fmt.Println(botWithMoves.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll+Str.
	// 2d6+Str => **(6 + 4) + -4** => **6**
	// The thing went terribly wrong. Mark XP.
}

func ExampleBot_MakeMove_noRoll() {
	bot := &Bot{
		db: &JsonDatabase{},
		moves: map[string]Move{
			"move": {
				Name:        "Move",
				Description: "When you do a thing, roll some dice.",
				Roll:        "",
				Hit:         "",
				Pass:        "",
				Miss:        "",
			},
		},
	}

	fmt.Println(bot.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll some dice.
}

func ExampleBot_MakeMove_noOutcomes() {
	rand.Seed(1)

	bot := &Bot{
		db: &JsonDatabase{},
		moves: map[string]Move{
			"move": {
				Name:        "Move",
				Description: "When you do a thing, roll some dice.",
				Roll:        "2d6",
				Hit:         "",
				Pass:        "",
				Miss:        "",
			},
		},
	}

	fmt.Println(bot.MakeMove(context, "Move"))
	// Output:
	// Player makes a move: Move!
	// When you do a thing, roll some dice.
	// 2d6 => **(6 + 4)** => **10**
}

func TestBot_HandleMessage_move(t *testing.T) {
	got := botWithMoves.HandleMessage(context, "!move Move")
	if strings.Index(got, "Player makes a move:") != 0 {
		t.Errorf("HandleMessage should make a move, got %v", got)
	}
}

func TestBot_HandleMessage_listMoves(t *testing.T) {
	got := botWithMoves.HandleMessage(context, "!move")
	if strings.Index(got, "I know the following moves") == -1 || strings.Index(got, "Move") == -1 {
		t.Errorf("HandleMessage should list moves, got %v", got)
	}
}
