package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type RetroCard struct {
	Id           string `json:"id"`
	ColumnId     string `json:"columnId"`
	Message      string `json:"message"`
	NumVotes     uint   `json:"numVotes"`
	IsEditable   bool   `json:"isEditable"`
	GroupId      string `json:"groupId"`
	IsDeleted    bool   `json:"isDeleted"`
	LastModified int    `json:"lastModified"`
}

func (r *RetroCard) UnmarshalJSON(data []byte) error {
	type target RetroCard

	if err := json.Unmarshal(data, (*target)(r)); err != nil {
		return err
	}

	const prefix = "-pk-"

	if !strings.Contains(r.Id, prefix) {
		return fmt.Errorf("got id '%s', expected it to contain '%s'", r.Id, prefix)
	}

	if r.Id == "" {
		return errors.New("id is empty")
	}

	if r.ColumnId == "" {
		return errors.New("column id is empty")
	}

	if r.Message == "" {
		return errors.New("message is empty")
	}

	if r.GroupId == "" {
		return errors.New("group id is empty")
	}

	if r.LastModified == 0 {
		return errors.New("last modified is empty")
	}

	r.IsEditable = false

	return nil
}
