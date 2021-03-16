package data

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Action struct {
	Title   string     `json:"title"`
	OldCard *RetroCard `json:"oldCard"`
	NewCard *RetroCard `json:"newCard"`
}

func (a *Action) UnmarshalJSON(data []byte) error {
	type target Action

	if err := json.Unmarshal(data, (*target)(a)); err != nil {
		return err
	}

	if a.Title != "upVote" {
		return fmt.Errorf("invalid action title '%s'", a.Title)
	}

	switch a.Title {
	case "upVote":
		if a.OldCard == nil {
			return errors.New("old card is nil")
		}

		if a.NewCard == nil {
			return errors.New("new card is nil")
		}
	}

	return nil
}
