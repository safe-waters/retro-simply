package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

var RoomIDRegex = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

type Room struct {
	Id       string `json:"id"`
	Password string `json:"password"`
}

func (r *Room) UnmarshalJSON(data []byte) error {
	type target Room

	if err := json.Unmarshal(data, (*target)(r)); err != nil {
		return err
	}

	if r.Password == "" {
		return PasswordInvalidError{errors.New("password cannot be empty")}
	}

	if !RoomIDRegex.MatchString(r.Id) {
		return RoomIdInvalidError{
			fmt.Errorf(
				"invalid room '%s' - it may contain letters, numbers, underscores and dashes",
				r.Id,
			),
		}
	}

	return nil
}
