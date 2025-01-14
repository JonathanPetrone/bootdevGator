package main

import (
	"fmt"
	configGator "github/jonathanpetrone/bootdevBlogAgg/internal/config"
)

func main() {
	c := &configGator.Config{}

	// Read the configuration
	err := configGator.ReadConfig(c)
	if err != nil {
		fmt.Println("Failed to read config:", err)
		return
	}

	// Set the user
	err = configGator.SetUser(c)
	if err != nil {
		fmt.Println("Failed to set user:", err)
		return
	}

	fmt.Printf("Final Config: %+v\n", c)
}
