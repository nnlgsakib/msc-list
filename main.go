package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Chain struct {
	Name           string         `json:"name"`
	Chain          string         `json:"chain"`
	Icon           string         `json:"icon"`
	RPC            []string       `json:"rpc"`
	Features       []Feature      `json:"features"`
	Faucets        []string       `json:"faucets"`
	NativeCurrency NativeCurrency `json:"nativeCurrency"`
	InfoURL        string         `json:"infoURL"`
	ShortName      string         `json:"shortName"`
	ChainID        int            `json:"chainId"`
	NetworkID      int            `json:"networkId"`
	ENS            *ENS           `json:"ens,omitempty"`       // Optional field
	Explorers      []Explorer     `json:"explorers,omitempty"` // Optional field
}

type Feature struct {
	Name string `json:"name"`
}

type NativeCurrency struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type ENS struct {
	Registry string `json:"registry"`
}

type Explorer struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Icon     string `json:"icon,omitempty"` // Optional field
	Standard string `json:"standard"`
}

// Define our own error to stop the walk
var ErrStopWalk = errors.New("stop walk")

func main() {
	// Directories for data
	infoFolder := "./data/info"
	iconFolder := "./data/icons"

	// Base URL for icons
	baseURL := "https://raw.githubusercontent.com/ethereum-lists/chains/refs/heads/master/_data/icons/"

	// Read JSON files from the info folder
	files, err := ioutil.ReadDir(infoFolder)
	if err != nil {
		log.Fatal(err)
	}

	var chains []Chain

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			jsonPath := filepath.Join(infoFolder, file.Name())

			// Read the JSON file
			jsonFile, err := os.Open(jsonPath)
			if err != nil {
				log.Println("Error opening JSON file:", err)
				continue
			}

			byteValue, _ := ioutil.ReadAll(jsonFile)
			jsonFile.Close()

			var chain Chain
			if err := json.Unmarshal(byteValue, &chain); err != nil {
				log.Println("Error parsing JSON file:", err)
				continue
			}

			// Update the icon field with the full URL
			if chain.Icon != "" {
				iconName := chain.Icon
				// Find the icon file with any allowed extension
				iconPath, err := findIconFile(iconFolder, iconName)
				if err != nil {
					log.Println(err)
				} else {
					extension := filepath.Ext(iconPath)
					// Construct the full URL using the actual extension
					chain.Icon = fmt.Sprintf("%s%s%s", baseURL, iconName, extension)
				}
			}

			chains = append(chains, chain)
		}
	}

	// Write the combined data to chains.json
	outputFile, err := os.Create("chains.json")
	if err != nil {
		log.Fatal("Cannot create output file:", err)
	}
	defer outputFile.Close()

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(chains); err != nil {
		log.Fatal("Cannot encode JSON:", err)
	}

	fmt.Println("chains.json has been created successfully.")
}

// findIconFile searches for the icon file with any allowed extension
func findIconFile(iconFolder, iconName string) (string, error) {
	allowedExtensions := []string{".png", ".jpeg", ".jpg", ".webp", ".svg"}

	var foundPath string
	err := filepath.Walk(iconFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			baseName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			if strings.EqualFold(baseName, iconName) {
				ext := filepath.Ext(info.Name())
				for _, allowedExt := range allowedExtensions {
					if strings.EqualFold(ext, allowedExt) {
						foundPath = path
						return ErrStopWalk // Use our own error to stop the walk
					}
				}
			}
		}
		return nil
	})

	if err != nil && err != ErrStopWalk {
		return "", err
	}
	if foundPath == "" {
		return "", fmt.Errorf("Icon file not found for %s", iconName)
	}
	return foundPath, nil
}
