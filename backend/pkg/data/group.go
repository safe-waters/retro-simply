package data

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Group struct {
	Id         string       `json:"id"`
	ColumnId   string       `json:"columnId"`
	IsEditable bool         `json:"isEditable"`
	Title      string       `json:"title"`
	RetroCards []*RetroCard `json:"retroCards"`
}

func (g *Group) UnmarshalJSON(data []byte) error {
	type target Group

	if err := json.Unmarshal(data, (*target)(g)); err != nil {
		return err
	}

	if g.Id == "" {
		return errors.New("id is empty")
	}

	if g.ColumnId == "" {
		return errors.New("column id is empty")
	}

	if g.Title == "" {
		return errors.New("title is empty")
	}

	if g.RetroCards == nil {
		return errors.New("retroCards is nil")
	}

	rIds := map[string]struct{}{}

	for _, r := range g.RetroCards {
		if r == nil {
			return errors.New("retroCard is nil")
		}

		if r.GroupId != g.Id {
			return fmt.Errorf("got group id '%s', expected '%s'", r.GroupId, g.Id)
		}

		if r.ColumnId != g.ColumnId {
			return fmt.Errorf(
				"got column id '%s', expected '%s'",
				r.ColumnId,
				g.ColumnId,
			)
		}

		if _, ok := rIds[r.Id]; ok {
			return fmt.Errorf("duplicate retro card id '%s' exists", r.Id)
		}

		rIds[r.Id] = struct{}{}
	}

	g.IsEditable = false

	return nil
}
