package data

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Column struct {
	Id        string     `json:"id"`
	Title     string     `json:"title"`
	CardStyle *CardStyle `json:"cardStyle"`
	Groups    []*Group   `json:"groups"`
}

func (c *Column) UnmarshalJSON(data []byte) error {
	type target Column

	if err := json.Unmarshal(data, (*target)(c)); err != nil {
		return err
	}

	if c.Id == "" {
		return errors.New("id is empty")
	}

	if c.Title == "" {
		return errors.New("title is empty")
	}

	if c.CardStyle == nil {
		return errors.New("card style is nil")
	}

	if len(c.Groups) < 1 {
		return errors.New("missing default group")
	}

	for _, g := range c.Groups {
		if g == nil {
			return errors.New("group is nil")
		}

		if g.ColumnId != c.Id {
			return fmt.Errorf(
				"got group column id '%s', expected '%s'",
				g.ColumnId,
				c.Id,
			)
		}
	}

	return nil
}
