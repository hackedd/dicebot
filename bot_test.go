package dicebot

import (
	"fmt"
	"math/rand"
	"testing"
)

var bot = &Bot{
	db: &JsonDatabase{},
}

var context = MessageContext{
	UserName:  "Player",
	UserId:    "user",
	ChannelId: "channel",
	ServerId:  "server",
}

func handleMessage(msg string) string {
	return bot.HandleMessage(context, msg)
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
	fmt.Println(handleMessage("!save 10 as ten"))
	fmt.Println(handleMessage("!roll ten"))
	// Output:
	// Saved **10** as `ten`
	// ten => **10**
}

func ExampleBot_HandleMessage_saveExpr() {
	rand.Seed(1)
	fmt.Println(handleMessage("!save 2d6 as r"))
	fmt.Println(handleMessage("!roll r+2"))
	fmt.Println(handleMessage("!roll r+2"))
	// Output:
	// Saved **2d6** as `r`
	// r+2 => **(6 + 4) + 2** => **12**
	// r+2 => **(6 + 6) + 2** => **14**
}

func ExampleBot_HandleMessage_saveCase() {
	fmt.Println(handleMessage("!save 10 as ten"))
	fmt.Println(handleMessage("!roll TEN"))
	fmt.Println(handleMessage("!save 5 as FIVE"))
	fmt.Println(handleMessage("!roll five"))
	// Output:
	// Saved **10** as `ten`
	// TEN => **10**
	// Saved **5** as `FIVE`
	// five => **5**
}

func ExampleBot_HandleMessage_saveFor() {
	fmt.Println(handleMessage("!save 5 as five for channel"))
	fmt.Println(handleMessage("!save 6 as six for server"))
	// Output:
	// Saved **5** as `five`
	// Saved **6** as `six`
}

func ExampleBot_HandleMessage_saveError() {
	fmt.Println(handleMessage("!save 10"))
	fmt.Println(handleMessage("!save 10 as x for y"))
	// Output:
	// Sorry, I don't understand how to parse 'save 10'
	// Sorry, I don't understand how to parse 'save 10 as x for y': undefined scope y
}

func TestBot_LookupVariable_Scope(t *testing.T) {
	bot.Save(context, "1", "v1", "user")
	_, err := bot.LookupVariable(context, "v1")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}

	otherUserContext := context
	otherUserContext.UserId = "another-user"

	otherChannelContext := context
	otherChannelContext.ChannelId = "another-channel"

	_, err = bot.LookupVariable(otherUserContext, "v1")
	if err == nil || err.Error() != "undefined variable `v1`" {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}

	bot.Save(context, "1", "v2", "channel")
	_, err = bot.LookupVariable(context, "v2")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
	_, err = bot.LookupVariable(otherUserContext, "v2")
	if err != nil {
		t.Errorf("Unexpected error looking up variable: %s", err)
	}
	_, err = bot.LookupVariable(otherChannelContext, "v2")
	if err == nil || err.Error() != "undefined variable `v2`" {
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
	// Sorry, I don't understand how to parse 'x': undefined variable `x`
}

func TestBot_HandleMessage_IgnoreUnknown(t *testing.T) {
	got := handleMessage("!foo")
	if got != "" {
		t.Errorf("HandleMessage should ignore unknown commands, got %v", got)
	}
}
