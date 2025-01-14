package configGator

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"
const userName = "Jonathan"

func SetUser(c *Config) error {
	c.Current_user_name = userName

	file, err := os.OpenFile(configFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("read")

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty print JSON with indentation

	// Write the struct as JSON
	err = encoder.Encode(c)
	if err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	fmt.Println("Config written to file successfully")
	return nil
}

func ReadConfig(c *Config) error {
	file, err := os.Open(configFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Println("read")

	defer file.Close()

	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return err
	}

	fmt.Printf("Config: %+v\n", c)

	return nil
}
