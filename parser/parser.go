package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Message represents a single chat message
type Message struct {
	ChatName    string
	Date        time.Time
	From        string
	Text        string
	Length      int
	IsReply     bool
	IsForwarded bool
}

// ChatMetadata contains information about the chat
type ChatMetadata struct {
	Name         string
	Type         string
	FirstMessage time.Time
	LastMessage  time.Time
	TotalCount   int
}

// ParseResult contains all parsed data
type ParseResult struct {
	Metadata ChatMetadata
	Messages []Message
}

// ParseAllFiles parses all messages*.html files in the given directory
func ParseAllFiles(dir string) (*ParseResult, error) {
	files, err := filepath.Glob(filepath.Join(dir, "messages*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no messages*.html files found in %s", dir)
	}

	// Sort files to process in order
	sort.Slice(files, func(i, j int) bool {
		return extractFileNumber(files[i]) < extractFileNumber(files[j])
	})

	result := &ParseResult{
		Messages: make([]Message, 0),
	}

	for i, file := range files {
		messages, chatName, err := parseFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}

		if i == 0 && chatName != "" {
			result.Metadata.Name = chatName
			result.Metadata.Type = "личный"
		}

		result.Messages = append(result.Messages, messages...)

		if (i+1)%20 == 0 {
			fmt.Printf("Обработано %d/%d файлов...\n", i+1, len(files))
		}
	}

	if len(result.Messages) > 0 {
		result.Metadata.FirstMessage = result.Messages[0].Date
		result.Metadata.LastMessage = result.Messages[len(result.Messages)-1].Date
		result.Metadata.TotalCount = len(result.Messages)
	}

	fmt.Printf("Всего обработано: %d файлов, %d сообщений\n", len(files), len(result.Messages))

	return result, nil
}

// extractFileNumber extracts the number from filename like "messages123.html"
func extractFileNumber(filename string) int {
	base := filepath.Base(filename)
	re := regexp.MustCompile(`messages(\d*)\.html`)
	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 || matches[1] == "" {
		return 0
	}
	var num int
	fmt.Sscanf(matches[1], "%d", &num)
	return num
}

// parseFile parses a single HTML file and returns messages
func parseFile(filename string) ([]Message, string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		return nil, "", err
	}

	// Extract chat name from header
	chatName := strings.TrimSpace(doc.Find(".page_header .text.bold").First().Text())

	messages := make([]Message, 0)
	var lastFrom string

	doc.Find(".message.default").Each(func(i int, s *goquery.Selection) {
		msg := Message{
			ChatName: chatName,
		}

		// Parse date from title attribute (get the first one, not from forwarded)
		dateEl := s.Find("> .body > .pull_right.date.details").First()
		if dateEl.Length() == 0 {
			dateEl = s.Find(".date.details").First()
		}
		dateStr, exists := dateEl.Attr("title")
		if exists {
			msg.Date = parseDate(dateStr)
		}

		// Parse sender - for "joined" messages, use the last known sender
		// Get direct child from_name, not from forwarded content
		fromEl := s.Find("> .body > .from_name").First()
		if fromEl.Length() == 0 {
			fromEl = s.Find(".from_name").First()
		}
		fromName := strings.TrimSpace(fromEl.Text())

		// Clean up forwarded message sender names (they include date)
		// Format: "Name  DD.MM.YYYY HH:MM:SS" or "Name DD.MM.YYYY HH:MM:SS"
		fromName = cleanForwardedName(fromName)

		if fromName != "" {
			lastFrom = fromName
		}
		msg.From = lastFrom

		// Check if it's a reply
		msg.IsReply = s.Find(".reply_to").Length() > 0

		// Check if it's forwarded
		msg.IsForwarded = s.Find(".forwarded").Length() > 0

		// Extract text (skip media-only messages)
		// Get text from main body, not from forwarded content
		textEl := s.Find("> .body > .text").First()
		if textEl.Length() == 0 {
			textEl = s.Find(".text").First()
		}
		if textEl.Length() > 0 {
			// Get text content, removing nested elements like links but keeping their text
			msg.Text = strings.TrimSpace(textEl.Text())
			msg.Length = len([]rune(msg.Text))
		}

		// Only add messages with text content
		if msg.Text != "" && !msg.Date.IsZero() {
			messages = append(messages, msg)
		}
	})

	return messages, chatName, nil
}

// cleanForwardedName removes date suffix from forwarded message sender names
// Handles formats like "Name  DD.MM.YYYY HH:MM:SS" or "Name DD.MM.YYYY HH:MM:SS"
func cleanForwardedName(name string) string {
	// Pattern: look for " DD.MM.YYYY" anywhere in the string
	datePattern := regexp.MustCompile(`\s+\d{2}\.\d{2}\.\d{4}.*$`)
	cleaned := datePattern.ReplaceAllString(name, "")
	return strings.TrimSpace(cleaned)
}

// parseDate parses date from format "01.09.2020 23:56:18 UTC+03:00"
func parseDate(dateStr string) time.Time {
	// Remove timezone suffix for simpler parsing
	dateStr = strings.TrimSpace(dateStr)

	// Try parsing with timezone
	layouts := []string{
		"02.01.2006 15:04:05 MST",
		"02.01.2006 15:04:05",
	}

	// Extract just the date and time part
	parts := strings.Split(dateStr, " UTC")
	if len(parts) > 0 {
		dateStr = strings.TrimSpace(parts[0])
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	// Try to parse just date and time
	if t, err := time.Parse("02.01.2006 15:04:05", dateStr); err == nil {
		return t
	}

	return time.Time{}
}
