package database

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"log"
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

func (p *Provider) InitMySQLDatabases(databases []MySQLConfiguration) error {
	for _, database := range databases {
		databaseUrl := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=True",
			database.Username,
			database.Password,
			database.Host,
			database.Database)

		connection, err := sql.Open("mysql", databaseUrl)
		if err != nil {
			return fmt.Errorf("could not connect to database %s %w", database.Name, err)
		}

		log.Println("Connected to database", database.Name)

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
