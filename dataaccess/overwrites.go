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
	slog.PrintDebug("Saving Overwrites")
	slog.PrintTrace("Overwrites being saved:", Overwrites)

	switch Backend {
		case "file":
			err := fileio.SaveOverwrites(Overwrites)
			slog.PrintTrace("Failed to write overwrites to disk:", err)
			return err
	}
	return errors.New("Selected backend is unknown")
}

func LoadOverwrites() error {
	slog.PrintDebug("Loading Overwrites")

	switch Backend {
		case "file":
			err := fileio.LoadOverwrites(&Overwrites)
			slog.PrintTrace("Failed to load overwrites from disk:", err)
			return err
	}
	return errors.New("Selected backend is unknown")
}
