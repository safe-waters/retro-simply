package data

import (
	"encoding/json"
	"errors"
)

type CardStyle struct {
	BackgroundColor string `json:"backgroundColor"`
}

func (c *CardStyle) UnmarshalJSON(data []byte) error {
	type target CardStyle

	if err := json.Unmarshal(data, (*target)(c)); err != nil {
		return err
	}

	if c.BackgroundColor == "" {
		return errors.New("background color is nil")
	}

	return nil
}
