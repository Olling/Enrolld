package dataaccess

import (
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/fileio"
)

var (
	Overwrites		map[string]objects.Overwrite
)

func SaveOverwrites() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Saving Overwrites")
			fileio.SaveOverwrites(Overwrites)
			return nil
	}
	return errors.New("Selected backend is unknown")
}

func LoadOverwrites() error {
	switch Backend {
		case "file":
			slog.PrintDebug("Loading Overwrites")
			fileio.LoadOverwrites(&Overwrites)
			return nil
	}
	return errors.New("Selected backend is unknown")
}
