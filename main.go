package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	configGator "github/jonathanpetrone/bootdevBlogAgg/internal/config"
	"github/jonathanpetrone/bootdevBlogAgg/internal/database"
	"os"
)

type state struct {
	db         *database.Queries
	configFile *configGator.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if handlerFunc, exists := c.handlers[cmd.name]; exists {
		err := handlerFunc(s, cmd)
		return err
	} else {
		return errors.New("Function command doesn't exist")
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("missing argument for login")
	}

	username := cmd.args[0]

	// Try to get the user
	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		// If no rows found, user doesn't exist
		if err == sql.ErrNoRows {
			fmt.Printf("User %s does not exist\n", username)
			os.Exit(1)
		}
		return err
	}

	s.configFile.Current_user_name = username

	err = configGator.WriteConfig(s.configFile)
	if err != nil {
		return err
	}

	fmt.Printf("The user %s has been set\n", cmd.args[0])

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("username required")
	}
	username := cmd.args[0]

	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)

	if err != nil {
		// Check if this is a unique violation error
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			fmt.Printf("User %s already exists\n", username)
			os.Exit(1)
		}
		return err
	}

	s.configFile.Current_user_name = username

	err = configGator.WriteConfig(s.configFile)
	if err != nil {
		return err
	}

	// User-friendly message
	fmt.Printf("User %s successfully created!\n", username)

	// Debug information
	fmt.Printf("User details: %+v\n", user)

	return nil
}

func main() {

	// initialize configfile
	c := &configGator.Config{}

	// Read the configuration
	err := configGator.ReadConfig(c)
	if err != nil {
		fmt.Println("Failed to read config:", err)
		return
	}

	db, err := sql.Open("postgres", c.Db_url)
	if err != nil {
		// handle error
	}

	dbQueries := database.New(db)

	// initialize state to hold configfile and dbqueries
	s := &state{
		configFile: c,
		db:         dbQueries,
	}

	// initialize handers
	cmds := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)

	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	err = cmds.run(s, cmd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Final Config: %+v\n", c)
}
