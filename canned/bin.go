package canned

import (
	"context"
	"encoding/json"

	"github.com/influxdata/chronograf"
)

//go:generate go-bindata -o bin_gen.go -ignore README|apps|.sh|go -pkg canned .

// BinLayoutsStore represents a layout store using data generated by go-bindata
type BinLayoutsStore struct {
	Logger chronograf.Logger
}

// All returns the set of all layouts
func (s *BinLayoutsStore) All(ctx context.Context) ([]chronograf.Layout, error) {
	names := AssetNames()
	layouts := make([]chronograf.Layout, len(names))
	for i, name := range names {
		octets, err := Asset(name)
		if err != nil {
			s.Logger.
				WithField("component", "apps").
				WithField("name", name).
				Error("Invalid Layout: ", err)
			return nil, chronograf.ErrLayoutInvalid
		}

		var layout chronograf.Layout
		if err = json.Unmarshal(octets, &layout); err != nil {
			s.Logger.
				WithField("component", "apps").
				WithField("name", name).
				Error("Unable to read layout:", err)
			return nil, chronograf.ErrLayoutInvalid
		}
		layouts[i] = layout
	}

	return layouts, nil
}

// Get retrieves Layout if `ID` exists.
func (s *BinLayoutsStore) Get(ctx context.Context, ID string) (chronograf.Layout, error) {
	layouts, err := s.All(ctx)
	if err != nil {
		s.Logger.
			WithField("component", "apps").
			WithField("name", ID).
			Error("Invalid Layout: ", err)
		return chronograf.Layout{}, chronograf.ErrLayoutInvalid
	}

	for _, layout := range layouts {
		if layout.ID == ID {
			return layout, nil
		}
	}

	s.Logger.
		WithField("component", "apps").
		WithField("name", ID).
		Error("Layout not found")
	return chronograf.Layout{}, chronograf.ErrLayoutNotFound
}
