package dicebot

import (
	"fmt"
	"math/rand"
	"testing"
)

var bot = &Bot{
	db: &JsonDatabase{},
}

func handleMessage(msg string) string {
	return bot.HandleMessage(msg, "server", "channel", "user")
}

func ExampleEscapeMarkdown() {
	fmt.Println(EscapeMarkdown("1 * 2 + `a`"))
	// Output: 1 \* 2 + \`a\`
}

func ExampleBot_HandleMessage_empty() {
	fmt.Println(handleMessage("!roll")[:25] + "...")
	// Output: Type `!roll d<x>` to roll...
}

func ExampleBot_HandleMessage_help() {
	fmt.Println(handleMessage("!roll help")[:25] + "...")
	// Output: Type `!roll d<x>` to roll...
}

func ExampleBot_HandleMessage_rollSimple() {
	rand.Seed(1)
	fmt.Println(handleMessage("!roll d6"))
	// Output: d6 => **6**
}

func ExampleBot_HandleMessage_roll() {
	rand.Seed(1)
	fmt.Println(handleMessage("!roll 2d6"))
	// Output: 2d6 => **(6 + 4)** => **10**
}

func ExampleBot_HandleMessage_save() {
	fmt.Println(handleMessage("!roll save 10 as ten"))
	fmt.Println(handleMessage("!roll ten"))
	// Output:
	// Saved **10** as `ten`
	// ten => **10**
}

func ExampleBot_HandleMessage_saveExpr() {
	rand.Seed(1)
	fmt.Println(handleMessage("!roll save 2d6 as r"))
	fmt.Println(handleMessage("!roll r+2"))
	fmt.Println(handleMessage("!roll r+2"))
	// Output:
	// Saved **2d6** as `r`
	// r+2 => **(6 + 4) + 2** => **12**
	// r+2 => **(6 + 6) + 2** => **14**
}

func ExampleBot_HandleMessage_saveFor() {
	fmt.Println(handleMessage("!roll save 5 as five for channel"))
	fmt.Println(handleMessage("!roll save 6 as six for server"))
	// Output:
	// Saved **5** as `five`
	// Saved **6** as `six`
}

func ExampleBot_HandleMessage_saveError() {
	fmt.Println(handleMessage("!roll save 10"))
	fmt.Println(handleMessage("!roll save 10 as x for y"))
	// Output:
	// Sorry, I don't understand how to parse 'save 10'
	// Sorry, I don't understand how to parse 'save 10 as x for y': Undefined scope y
}

func TestBot_LookupVariable_Scope(t *testing.T) {
	bot.Save("1", "v1", "user-user")
	_, err := bot.LookupVariable("v1", "user", "channel", "server")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
	_, err = bot.LookupVariable("v1", "another-user", "channel", "server")
	if err == nil || err.Error() != "Undefined variable `v1`" {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}

	bot.Save("1", "v2", "channel-channel")
	_, err = bot.LookupVariable("v2", "user", "channel", "server")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
	_, err = bot.LookupVariable("v2", "another-user", "channel", "server")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
	_, err = bot.LookupVariable("v2", "user", "another-channel", "server")
	if err == nil || err.Error() != "Undefined variable `v2`" {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
}

func ExampleBot_HandleMessage_error1() {
	fmt.Println(handleMessage("!roll 1.5d6"))
	// Output:
	// Sorry, I don't understand how to parse '1.5d6'
	// ```
	// 1.5d6
	//  ^-- Input not matched
	// ```
}

func ExampleBot_HandleMessage_error2() {
	fmt.Println(handleMessage("!roll 3**3"))
	// Output:
	// Sorry, I don't understand how to parse '3\*\*3'
	// ```
	// 3**3
	//   ^-- Unexpected input
	// ```
}

func ExampleBot_HandleMessage_error3() {
	fmt.Println(handleMessage("!roll x"))
	// Output:
	// Sorry, I don't understand how to parse 'x': Undefined variable `x`
}
