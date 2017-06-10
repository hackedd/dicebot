package dicebot

import (
	"fmt"
	"math/rand"
)

var bot = &Bot{}

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
