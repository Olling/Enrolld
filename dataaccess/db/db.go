package db

import (
	"log"

	"github.com/Olling/Enrolld/dataaccess/config"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/slog"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

// GetDbConnection establishes the database connection
func GetDbConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", "postgresql://"+config.Configuration.DBUser+":"+config.Configuration.DBPass+"@"+config.Configuration.DBHost+":"+config.Configuration.DBPort+"/"+config.Configuration.DBInstance)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// MigrateDB makes sure the DB exists and migrates it to the latest version
func MigrateDB() {
	db, err := GetDbConnection()
	if err != nil {
		slog.PrintFatal("Could not connect to db:", err)
		return
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://./dataaccess/db/migrations",
		"postgres", driver)

	if err != nil {
		log.Fatal("DB migration conn failed:", err)
	}

	err = m.Up()
	if err != nil {
		log.Println("DB migration failed:", err)
	}

}

func LoadAuthentication(users *map[string]objects.User) error {
	return nil
}

func DeleteServer(serverName string) error {
	return nil
}

func LoadOverwrites(overwrites interface{}) error {
	//Write to overwrites
	return nil
}

func SaveOverwrites(overwrites interface{}) error {
	return nil
}

func AddOverwrites(server *objects.Server, overwrites map[string]objects.Overwrite) {
}

// GetServers returns a list of all servers
func GetServers(overwrites map[string]objects.Overwrite) (servers []objects.Server, err error) {
	db, err := GetDbConnection()
	if err != nil {
		slog.PrintFatal("Could not connect to db:", err)
		return
	}
	defer db.Close()

	err = db.Get(&servers, "SELECT * FROM servers")
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// ServerExist checks if a specific sever exists based on its ID
func ServerExist(serverID string) bool {
	db, err := GetDbConnection()
	if err != nil {
		slog.PrintFatal("Could not connect to db:", err)
		return false
	}
	defer db.Close()

	var exists bool
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM ...)")
	row.Scan(&exists)
	return exists
}

// RemoveServer removes a server from the DB based on its ID
func RemoveServer(serverID string) error {
	//if err == nil {
	//	metrics.ServersDeleted.Inc()
	//}
	db, err := GetDbConnection()
	if err != nil {
		slog.PrintFatal("Could not connect to db:", err)
		return nil
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM servers WHERE id = $1", serverID)
	if err != nil {
		slog.PrintError("Could not delete server", serverID, " Error: ", err)
	}

	return nil
}

func GetServer(serverID string, overwrites map[string]objects.Overwrite) (server objects.Server, err error) {
	return server, err
}

func UpdateServer(server objects.Server, isNewServer bool) error {
	return nil
}

func IsServerActive(serverID string) bool {
	return false
}

func MarkServerActive(server objects.Server) error {
	return nil
}

func MarkServerInactive(server objects.Server) error {
	return nil
}

func GetServerCount() float64 {
	return 0
}

func GetFilteredServersList(groups []string, properties map[string]string) ([]objects.Server, error) {
	return []objects.Server{}, nil
}
