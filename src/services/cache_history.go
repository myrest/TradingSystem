package services

import (
	"TradingSystem/src/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	cacheMu  sync.Mutex
	cacheDir string
)

func getCacheFilePath(key string) string {
	return filepath.Join(cacheDir, fmt.Sprintf("%s.json", key))
}

func getLog_TVCacheKey(CustomerID, Symbol, DaysAgo string) string {
	return fmt.Sprintf("%s_%s_%s", CustomerID, Symbol, DaysAgo)
}

func saveLog_TVCache(key string, data []models.Log_TvSiginalData) error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	filePath := getCacheFilePath(key)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func loadLog_TVCache(key string) ([]models.Log_TvSiginalData, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	filePath := getCacheFilePath(key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []models.Log_TvSiginalData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	return data, err
}

func saveDemoSymbolListCache(key string, data []models.DemoSymbolList) error {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	filePath := getCacheFilePath(key)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func loadDemoSymbolListCache(key string) ([]models.DemoSymbolList, error) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	filePath := getCacheFilePath(key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []models.DemoSymbolList
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	return data, err
}
