package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	redisClient *redis.Client
}

func (s *Service) run(ctx context.Context) error {
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
			msg, err := s.check(ctx, domain)
			if err != nil {
				return err
			}

			log.Print(msg)
		}
	}

	return nil
}

func (s *Service) check(ctx context.Context, domain string) (string, error) {
	val, err := s.redisClient.Get(ctx, domain).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	if err == nil {
		return val + " (cached)", nil
	}

	// not cached
	msg, err := s.lookup(domain)
	if err != nil {
		return "", err
	}

	err = s.redisClient.Set(ctx, domain, string(msg), 24*time.Hour).Err()
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (s *Service) lookup(domain string) (string, error) {
	raw, err := whois.Whois(domain)

	result, err := whoisparser.Parse(raw)
	if errors.Is(err, whoisparser.ErrNotFoundDomain) {
		msg := fmt.Sprintf("✅ %s is available", domain)
		return msg, nil
	}

	if err != nil {
		return "", err
	}

	hasAvailableStatus := false
	for _, status := range result.Domain.Status {
		if status == "available" {
			hasAvailableStatus = true
		}
	}

	if hasAvailableStatus {
		if len(result.Domain.Status) > 1 {
			msg := fmt.Sprintf("❓ %s may be available", domain)
			return msg, nil
		}

		msg := fmt.Sprintf("✅ %s is available", domain)
		return msg, nil
	}

	// 10 days
	targetTime := time.Now().Add(24 * time.Hour * 10)
	if result.Domain.ExpirationDateInTime.Before(targetTime) {
		msg := fmt.Sprintf("ℹ️  %s may expire soon (%s)", domain, result.Domain.ExpirationDate)
		return msg, nil
	}

	msg := fmt.Sprintf("❌ %s is not available", domain)
	return msg, nil
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
