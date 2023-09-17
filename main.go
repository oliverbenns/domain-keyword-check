package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

type Word struct {
	Text     string `json:"text"`
	Location int    `json:"location"`
}

type Data struct {
	Words []Word `json:"words"`
}

func main() {
	err := run()
	if err != nil {
		log.Panic(err.Error())
	}
}

func run() error {
	file, err := os.Open("data.json")
	if err != nil {
		return err
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	data := Data{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	firstWords := filterByLocation(data.Words, 0)
	secondWords := filterByLocation(data.Words, 1)

	for _, firstWord := range firstWords {
		for _, secondWord := range secondWords {
			domain := fmt.Sprintf("%s%s.com", firstWord.Text, secondWord.Text)
			err := check(domain)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func check(domain string) error {
	raw, err := whois.Whois(domain)

	result, err := whoisparser.Parse(raw)
	if errors.Is(err, whoisparser.ErrNotFoundDomain) {
		log.Printf("✅ %s is available", domain)
		return nil
	}

	if err != nil {
		return err
	}

	hasAvailableStatus := false
	for _, status := range result.Domain.Status {
		if status == "available" {
			hasAvailableStatus = true
		}
	}

	if hasAvailableStatus {
		if len(result.Domain.Status) > 1 {
			log.Printf("❓ %s may be available", domain)
		} else {
			log.Printf("✅ %s is available", domain)
		}
		return nil
	}

	// 10 days
	targetTime := time.Now().Add(24 * time.Hour * 10)
	if result.Domain.ExpirationDateInTime.Before(targetTime) {
		log.Printf("ℹ️  %s may expire soon", domain)
		return nil
	}

	log.Printf("❌ %s is not available", domain)
	return nil
}

func filterByLocation(words []Word, loc int) []Word {
	filteredWords := []Word{}
	for _, word := range words {
		if word.Location == loc {
			filteredWords = append(filteredWords, word)
		}
	}

	return filteredWords
}
