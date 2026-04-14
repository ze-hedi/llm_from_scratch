package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type SaveData struct {
	HighScore int `json:"high_score"`
}

func getSaveFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	saveDir := filepath.Join(homeDir, ".cli_go", "dino")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(saveDir, "save.json"), nil
}

func LoadHighScore() int {
	path, err := getSaveFilePath()
	if err != nil {
		return 0
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return 0
	}

	return saveData.HighScore
}

func SaveHighScore(highScore int) error {
	path, err := getSaveFilePath()
	if err != nil {
		return err
	}

	saveData := SaveData{
		HighScore: highScore,
	}

	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
