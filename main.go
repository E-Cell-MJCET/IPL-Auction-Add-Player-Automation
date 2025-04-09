package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/xuri/excelize/v2"
)

type Player struct {
	PlayerName  string  `json:"playerName"`
	PlayerId    string  `json:"playerId"`
	Rating      float64 `json:"rating"`
	BoughtAt    any     `json:"boughtAt"`
	BasePrice   int     `json:"basePrice"`
	Pocket      string  `json:"pocket"`
	Nationality string  `json:"nationality"`
	Role        string  `json:"role"`
}

func generatePlayerId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("PL%04d", rand.Intn(10000))
}

func formatPool(poolValue string) string {
	if len(poolValue) > 1 && poolValue[0] == 'P' {
		return poolValue[1:]
	}
	return poolValue
}

func readExcelFile(filePath string) []Player {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		log.Fatalf("Error reading rows: %v", err)
	}

	var players []Player
	maxPlayers := 10
	endIndex := len(rows[1:])
	if endIndex > maxPlayers {
		endIndex = maxPlayers
	}

	for _, row := range rows[1 : endIndex+1] {
		// if len(row) < 7 {
		// 	log.Printf("Skipping row with insufficient data: %v", row)
		// 	continue
		// }

		player := Player{
			PlayerName:  row[0],
			PlayerId:    generatePlayerId(),               // Using generated ID instead of reading from Excel
			Rating:      parseFloatOrDefault(row[1], 0.0), // Fixed index for rating
			BoughtAt:    nil,
			BasePrice:   parseIntOrDefault(row[5], 0),
			Pocket:      formatPool(row[2]), // Format the pool value
			Nationality: parseNationality(row[4]),
			Role:        row[3],
		}
		players = append(players, player)
	}
	return players
}

func parseIntOrDefault(s string, defaultValue int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue
	}
	return result
}

func parseNationality(s string) string {
	if s == "India" {
		return "Indian"
	} else {
		return "Foreign"
	}
}

func parseFloatOrDefault(s string, defaultValue float64) float64 {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	if err != nil {
		return defaultValue
	}
	return result
}



func sendPlayerToServer(player Player) error {
	jsonData, err := json.Marshal(player)
	if err != nil {
		return fmt.Errorf("error marshaling player data: %v", err)
	}

	resp, err := http.Post("http://localhost:8080/api/player", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	// json format player ID and name is stored in json file and all players are appended one by one

	appendToJSONFile(player)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %v", resp.Status)
	}
	return nil
}

// appendToJSONFile appends player name and player ID to a JSON file
func appendToJSONFile(player Player) error {
	// Define the file path for storing player data
	filePath := "players.json"

	// Open the file for reading and appending
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening JSON file: %v", err)
	}
	defer file.Close()

	// Create a map to store player data (name and ID)
	playerData := map[string]string{
		"playerName": player.PlayerName,
		"playerId":   player.PlayerId,
	}

	// Convert the player data map to JSON
	data, err := json.Marshal(playerData)
	if err != nil {
		return fmt.Errorf("error marshaling player data: %v", err)
	}

	// Add a newline before appending to separate entries
	data = append(data, '\n')

	// Write the player data to the file
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("error writing to JSON file: %v", err)
	}

	return nil
}

func main() {
	filePath := "ipl-sheet.xlsx"
	players := readExcelFile(filePath)

	for _, player := range players {
		if err := sendPlayerToServer(player); err != nil {
			log.Printf("Error sending player %s: %v", player.PlayerName, err)
			continue
		}
		log.Printf("Successfully sent player: %s", player.PlayerName)
	}
}
