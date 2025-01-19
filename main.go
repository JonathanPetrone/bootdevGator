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

func handlerResetUsers(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Println("Failed to reset users", err)
		return err
	}

	s.configFile.Current_user_name = ""

	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Failed to get users", err)
		return err
	}

	for user := range users {
		if users[user].Name == s.configFile.Current_user_name {
			fmt.Printf("* %s (current)\n", users[user].Name)
		} else {
			fmt.Printf("* %s\n", users[user].Name)
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	ctx := context.Background()
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(ctx, url)
	if err != nil {
		return err
	}
	// Print the feed struct somehow
	fmt.Printf("Struct contents: %+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("expected 2 arguments: name and url")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	current_user := s.configFile.Current_user_name
	user, err := s.db.GetUser(context.Background(), current_user)
	if err != nil {
		return err
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return err
	}

	follows, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed follow: %v", err)
	}

	fmt.Printf("Feed created successfully:\n")
	fmt.Printf("  Name: %s\n", feed.Name)
	fmt.Printf("  URL: %s\n", feed.Url)
	fmt.Printf("  ID: %s\n", feed.ID)

	fmt.Printf("  Following: %s\n", follows[0].FeedName)

	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("List of Feeds:")
	for i := range feeds {
		fmt.Printf("  Name: %s\n", feeds[i].Name)
		fmt.Printf("  URL: %s\n", feeds[i].Url)
		fmt.Printf("  Added by: %s\n", feeds[i].Username)
	}

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("expected 1 arguments: url")
	}

	currentUser, err := s.db.GetUser(context.Background(), s.configFile.Current_user_name) // assuming Name is stored in config
	if err != nil {
		return fmt.Errorf("couldn't get current user: %v", err)
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("couldn't get feed: %v", err)
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})

	if err != nil {
		return fmt.Errorf("couldn't create follow: %v", err)
	}

	if len(follow) > 0 {
		fmt.Printf("Followed feed '%v' for user '%v'\n", follow[0].FeedName, follow[0].UserName)
	}

	return nil
}

func handlerFollowing(s *state, cmd command) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("following command takes no arguments")
	}

	currentUser, err := s.db.GetUser(context.Background(), s.configFile.Current_user_name) // assuming Name is stored in config
	if err != nil {
		return fmt.Errorf("couldn't get current user: %v", err)
	}

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("couldn't get follows: %v", err)
	}

	for _, follow := range follows {
		fmt.Printf("%v\n", follow.FeedName)
	}

	return nil
}

func main() {

	// initialize configfile
	c := &configGator.Config{}

	// Read the configuration
	err := configGator.ReadConfig(c)
	if err != nil {
		fmt.Println("Failed to read config:", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", c.Db_url)
	if err != nil {
		fmt.Println("Failed to connect to database:", err)
		os.Exit(1)
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
	cmds.register("reset", handlerResetUsers)
	cmds.register("users", handlerGetUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerGetFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)

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

	//fmt.Printf("Final Config: %+v\n", c)
}
