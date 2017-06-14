package dicebot

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

var bot *Bot

func TestMain(m *testing.M) {
	var err error

	bot, err = NewBot("test.db")
	if err != nil {
		fmt.Printf("Unable to open database: %s\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func ExampleEscapeMarkdown() {
	fmt.Println(EscapeMarkdown("1 * 2 + `a`"))
	// Output: 1 \* 2 + \`a\`
}

func ExampleBot_HandleMessage_empty() {
	fmt.Println(bot.HandleMessage("!roll")[:25] + "...")
	// Output: Type `!roll d<x>` to roll...
}

func ExampleBot_HandleMessage_help() {
	fmt.Println(bot.HandleMessage("!roll help")[:25] + "...")
	// Output: Type `!roll d<x>` to roll...
}

func ExampleBot_HandleMessage_rollSimple() {
	rand.Seed(1)
	fmt.Println(bot.HandleMessage("!roll d6"))
	// Output: d6 => **6**
}

func ExampleBot_HandleMessage_roll() {
	rand.Seed(1)
	fmt.Println(bot.HandleMessage("!roll 2d6"))
	// Output: 2d6 => **(6 + 4)** => **10**
}

func ExampleBot_HandleMessage_save() {
	fmt.Println(bot.HandleMessage("!roll save 10 as ten"))
	fmt.Println(bot.HandleMessage("!roll ten"))
	// Output:
	// Saved **10** as `ten`
	// ten => **10**
}

func ExampleBot_HandleMessage_saveExpr() {
	rand.Seed(1)
	fmt.Println(bot.HandleMessage("!roll save 2d6 as r"))
	fmt.Println(bot.HandleMessage("!roll r+2"))
	fmt.Println(bot.HandleMessage("!roll r+2"))
	// Output:
	// Saved **2d6** as `r`
	// r+2 => **(6 + 4) + 2** => **12**
	// r+2 => **(6 + 6) + 2** => **14**
}

func ExampleBot_HandleMessage_saveError() {
	fmt.Println(bot.HandleMessage("!roll save 10"))
	// Output:
	// Sorry, I don't understand how to parse 'save 10'
}

func ExampleBot_HandleMessage_error1() {
	fmt.Println(bot.HandleMessage("!roll 1.5d6"))
	// Output:
	// Sorry, I don't understand how to parse '1.5d6'
	// ```
	// 1.5d6
	//  ^-- Input not matched
	// ```
}

func ExampleBot_HandleMessage_error2() {
	fmt.Println(bot.HandleMessage("!roll 3**3"))
	// Output:
	// Sorry, I don't understand how to parse '3\*\*3'
	// ```
	// 3**3
	//   ^-- Unexpected input
	// ```
}

func ExampleBot_HandleMessage_error3() {
	fmt.Println(bot.HandleMessage("!roll x"))
	// Output:
	// Sorry, I don't understand how to parse 'x': Undefined variable `x`
}
