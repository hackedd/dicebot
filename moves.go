package dicebot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type Move struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Roll        string `json:"roll"`
	Hit         string `json:"hit"`
	Pass        string `json:"pass"`
	Miss        string `json:"miss"`
}

func LoadMoves(moves map[string]Move, filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var moveList []Move
	err = json.Unmarshal(data, &moveList)
	if err != nil {
		return err
	}

	for _, move := range moveList {
		moves[strings.ToLower(move.Name)] = move
	}

	return nil
}

func (bot *Bot) MakeMove(context MessageContext, moveName string) (output string) {
	move, ok := bot.moves[strings.ToLower(moveName)]
	if !ok {
		return bot.HandleError(moveName, errors.New("unknown move"))
	}

	output = fmt.Sprintf("%s makes a move: %s!\n", context.UserName, move.Name)
	output += move.Description + "\n"
	if move.Roll != "" {
		value, explanation, err := bot.Eval(context, move.Roll)
		if err != nil {
			output += bot.HandleError(move.Roll, err)
			return
		}

		output += bot.FormatResult(move.Roll, value, explanation) + "\n"

		if value >= 10 && move.Hit != "" {
			output += move.Hit + "\n"
		}
		if value >= 7 && value <= 9 && move.Pass != "" {
			output += move.Pass + "\n"
		}
		if value < 7 {
			if move.Miss != "" {
				output += move.Miss + " "
			}
			output += "Mark XP.\n"
		}
	}
	return
}
