package data

import (
	"encoding/json"
	"errors"
	"fmt"
)

type State struct {
	RoomId  string    `json:"roomId"`
	Columns []*Column `json:"columns"`
	Action  *Action   `json:"action"`
}

func (s *State) UnmarshalJSON(data []byte) error {
	type target State

	if err := json.Unmarshal(data, (*target)(s)); err != nil {
		return err
	}

	if s.RoomId == "" {
		return errors.New("room id is empty")
	}

	const numCols = 3

	if len(s.Columns) != numCols {
		return fmt.Errorf(
			"got '%d' columns, expected '%d'",
			len(s.Columns),
			numCols,
		)
	}

	for _, c := range s.Columns {
		if c == nil {
			return errors.New("column is nil")
		}
	}

	return nil
}
