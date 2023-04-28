package database

import (
	"context"
	"csrvbot/pkg/logger"
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
)

type Provider struct {
	databases map[string]*gorp.DbMap
}

func NewProvider() *Provider {
	return &Provider{
		databases: make(map[string]*gorp.DbMap),
	}
}

type MySQLConfiguration struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Database string `json:"database"`
}

func (p *Provider) InitMySQLDatabases(ctx context.Context, databases []MySQLConfiguration) error {
	log := logger.GetLoggerFromContext(ctx)
	for _, database := range databases {
		databaseUrl := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True",
			database.Username,
			database.Password,
			database.Host,
			database.Database)
		log.WithField("dbname", database.Name).Debug("Connecting to database")

		connection, err := sql.Open("mysql", databaseUrl)
		if err != nil {
			return fmt.Errorf("could not open database %s %w", database.Name, err)
		}

		if err := connection.Ping(); err != nil {
			return fmt.Errorf("could not ping database %s %w", database.Name, err)
		}

		log.WithField("dbname", database.Name).Info("Connected to database")

		p.databases[database.Name] = &gorp.DbMap{Db: connection, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8MB4"}}

	}

	return nil
}

func (p *Provider) GetMySQLDatabase(name string) (*gorp.DbMap, error) {
	database, ok := p.databases[name]
	if !ok {
		return nil, fmt.Errorf("could not find database %s", name)
	}
	return database, nil
}

func (p *Provider) CreateTablesIfNotExists() error {
	for name, database := range p.databases {
		err := database.CreateTablesIfNotExists()
		if err != nil {
			return fmt.Errorf("could not create tables in db %s %w", name, err)
		}
	}

	return nil
}
