package dataaccess

import (
	"errors"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/db"
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
		case "db":
			err := db.SaveOverwrites(Overwrites)
			slog.PrintTrace("Failed to write overwrites to database:", err)
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
		case "db":
			err := db.LoadOverwrites(&Overwrites)
			slog.PrintTrace("Failed to load overwrites from database:", err)
			return err
	}
	return errors.New("Selected backend is unknown")
}
