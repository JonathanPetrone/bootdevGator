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

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homeDir + "/.gatorconfig.json", nil
}

func ReadConfig(c *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Try to open the file
	file, err := os.Open(configPath)
	if err != nil {
		// If the file doesn't exist, create it with default values
		if os.IsNotExist(err) {
			// Initialize with default values
			c.Db_url = ""
			c.Current_user_name = ""
			// Write the default config
			return WriteConfig(c)
		}
		return err
	}
	defer file.Close()

	// If file exists, decode it
	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		return err
	}

	return nil
}

func WriteConfig(c *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
