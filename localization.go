package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// translations Keeps loaded transfers
var translations map[string]string

// loadTranslations Uploads a json file with translations
func loadTranslations(lang string) {
	filePath := fmt.Sprintf("locales/%s.json", lang)
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error loading localization (%s): %v", lang, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing localization (%s): %v", lang, err)
		}
	}(file)

	decoder := json.NewDecoder(file)
	translations = make(map[string]string)
	if err := decoder.Decode(&translations); err != nil {
		log.Fatalf("JSON parsing error (%s): %v", lang, err)
	}
	log.Printf("Localization loaded: %s", lang)
}

// t returns the translated line or the key itself if there is no translation
func t(key string, args ...interface{}) string {
	if msg, exists := translations[key]; exists {
		return fmt.Sprintf(msg, args...)
	}
	return fmt.Sprintf(key, args...) // Return the key if there is no translation
}
