package analyzer

import (
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"telegram_message_analyzer/parser"
	"telegram_message_analyzer/stopwords"
)

// YearStats contains statistics for a single year
type YearStats struct {
	Year                int
	TotalMessages       int
	MessagesByUser      map[string]int
	WordFrequency       map[string]int
	WordFrequencyByUser map[string]map[string]int // user -> word -> count
	TopWords            []WordCount
	TopWordsByUser      map[string][]WordCount // user -> top words
	HourlyActivity      map[int]int            // hour -> count
	MonthlyActivity     map[string]int         // "YYYY-MM" -> count
	MostActiveWindow    TimeWindow
	MostActiveMonth     MonthStat
	FirstMessage        time.Time
	LastMessage         time.Time
	RepliesCount        int
	ForwardedCount      int
	AvgMessageLength    float64
}

// WordCount represents a word with its count
type WordCount struct {
	Word  string
	Count int
}

// TimeWindow represents a 2-hour time window
type TimeWindow struct {
	StartHour int
	EndHour   int
	Count     int
}

// MonthStat represents month activity
type MonthStat struct {
	Year  int
	Month time.Month
	Count int
}

// Stats contains all analysis results
type Stats struct {
	Overall  YearStats
	ByYear   map[int]*YearStats
	ChatName string
	ChatType string
}

// Analyze performs full analysis on parsed messages
func Analyze(result *parser.ParseResult) *Stats {
	stats := &Stats{
		ByYear:   make(map[int]*YearStats),
		ChatName: result.Metadata.Name,
		ChatType: result.Metadata.Type,
		Overall: YearStats{
			MessagesByUser:      make(map[string]int),
			WordFrequency:       make(map[string]int),
			WordFrequencyByUser: make(map[string]map[string]int),
			HourlyActivity:      make(map[int]int),
			MonthlyActivity:     make(map[string]int),
		},
	}

	// Process each message
	for _, msg := range result.Messages {
		year := msg.Date.Year()

		// Initialize year stats if needed
		if _, ok := stats.ByYear[year]; !ok {
			stats.ByYear[year] = &YearStats{
				Year:                year,
				MessagesByUser:      make(map[string]int),
				WordFrequency:       make(map[string]int),
				WordFrequencyByUser: make(map[string]map[string]int),
				HourlyActivity:      make(map[int]int),
				MonthlyActivity:     make(map[string]int),
			}
		}

		yearStats := stats.ByYear[year]

		// Count messages
		yearStats.TotalMessages++
		stats.Overall.TotalMessages++

		// Count by user
		yearStats.MessagesByUser[msg.From]++
		stats.Overall.MessagesByUser[msg.From]++

		// Count replies and forwards
		if msg.IsReply {
			yearStats.RepliesCount++
			stats.Overall.RepliesCount++
		}
		if msg.IsForwarded {
			yearStats.ForwardedCount++
			stats.Overall.ForwardedCount++
		}

		// Track message length
		yearStats.AvgMessageLength += float64(msg.Length)
		stats.Overall.AvgMessageLength += float64(msg.Length)

		// Track first/last messages
		if yearStats.FirstMessage.IsZero() || msg.Date.Before(yearStats.FirstMessage) {
			yearStats.FirstMessage = msg.Date
		}
		if yearStats.LastMessage.IsZero() || msg.Date.After(yearStats.LastMessage) {
			yearStats.LastMessage = msg.Date
		}

		// Hourly activity
		hour := msg.Date.Hour()
		yearStats.HourlyActivity[hour]++
		stats.Overall.HourlyActivity[hour]++

		// Monthly activity
		monthKey := msg.Date.Format("2006-01")
		yearStats.MonthlyActivity[monthKey]++
		stats.Overall.MonthlyActivity[monthKey]++

		// Word frequency (overall and by user)
		words := extractWords(msg.Text)
		for _, word := range words {
			if !stopwords.IsStopWord(word) && len([]rune(word)) > 1 {
				yearStats.WordFrequency[word]++
				stats.Overall.WordFrequency[word]++

				// By user - year stats
				if yearStats.WordFrequencyByUser[msg.From] == nil {
					yearStats.WordFrequencyByUser[msg.From] = make(map[string]int)
				}
				yearStats.WordFrequencyByUser[msg.From][word]++

				// By user - overall stats
				if stats.Overall.WordFrequencyByUser[msg.From] == nil {
					stats.Overall.WordFrequencyByUser[msg.From] = make(map[string]int)
				}
				stats.Overall.WordFrequencyByUser[msg.From][word]++
			}
		}
	}

	// Calculate averages and top stats
	for _, yearStats := range stats.ByYear {
		if yearStats.TotalMessages > 0 {
			yearStats.AvgMessageLength /= float64(yearStats.TotalMessages)
		}
		yearStats.TopWords = getTopWords(yearStats.WordFrequency, 20)
		yearStats.TopWordsByUser = make(map[string][]WordCount)
		for user, wordFreq := range yearStats.WordFrequencyByUser {
			yearStats.TopWordsByUser[user] = getTopWords(wordFreq, 20)
		}
		yearStats.MostActiveWindow = getMostActiveWindow(yearStats.HourlyActivity)
		yearStats.MostActiveMonth = getMostActiveMonth(yearStats.MonthlyActivity)
	}

	if stats.Overall.TotalMessages > 0 {
		stats.Overall.AvgMessageLength /= float64(stats.Overall.TotalMessages)
	}
	stats.Overall.TopWords = getTopWords(stats.Overall.WordFrequency, 20)
	stats.Overall.TopWordsByUser = make(map[string][]WordCount)
	for user, wordFreq := range stats.Overall.WordFrequencyByUser {
		stats.Overall.TopWordsByUser[user] = getTopWords(wordFreq, 20)
	}
	stats.Overall.MostActiveWindow = getMostActiveWindow(stats.Overall.HourlyActivity)
	stats.Overall.MostActiveMonth = getMostActiveMonth(stats.Overall.MonthlyActivity)
	stats.Overall.FirstMessage = result.Metadata.FirstMessage
	stats.Overall.LastMessage = result.Metadata.LastMessage

	return stats
}

// extractWords extracts and normalizes words from text
func extractWords(text string) []string {
	// Remove URLs
	urlRegex := regexp.MustCompile(`https?://\S+`)
	text = urlRegex.ReplaceAllString(text, "")

	// Remove special characters but keep Cyrillic and Latin letters
	var result []string
	var currentWord strings.Builder

	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				word := currentWord.String()
				// Filter out pure numbers and very short words
				if len([]rune(word)) > 1 && !isNumber(word) {
					result = append(result, word)
				}
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		word := currentWord.String()
		if len([]rune(word)) > 1 && !isNumber(word) {
			result = append(result, word)
		}
	}

	return result
}

// isNumber checks if string contains only digits
func isNumber(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// getTopWords returns top N words by frequency
func getTopWords(freq map[string]int, n int) []WordCount {
	words := make([]WordCount, 0, len(freq))
	for word, count := range freq {
		words = append(words, WordCount{Word: word, Count: count})
	}

	sort.Slice(words, func(i, j int) bool {
		return words[i].Count > words[j].Count
	})

	if len(words) > n {
		words = words[:n]
	}

	return words
}

// getMostActiveWindow finds the most active 2-hour window
func getMostActiveWindow(hourly map[int]int) TimeWindow {
	windows := []TimeWindow{
		{0, 2, 0}, {2, 4, 0}, {4, 6, 0}, {6, 8, 0},
		{8, 10, 0}, {10, 12, 0}, {12, 14, 0}, {14, 16, 0},
		{16, 18, 0}, {18, 20, 0}, {20, 22, 0}, {22, 24, 0},
	}

	for i := range windows {
		for h := windows[i].StartHour; h < windows[i].EndHour; h++ {
			windows[i].Count += hourly[h]
		}
	}

	var maxWindow TimeWindow
	for _, w := range windows {
		if w.Count > maxWindow.Count {
			maxWindow = w
		}
	}

	return maxWindow
}

// getMostActiveMonth finds the most active month
func getMostActiveMonth(monthly map[string]int) MonthStat {
	var maxStat MonthStat

	for monthKey, count := range monthly {
		if count > maxStat.Count {
			t, err := time.Parse("2006-01", monthKey)
			if err == nil {
				maxStat = MonthStat{
					Year:  t.Year(),
					Month: t.Month(),
					Count: count,
				}
			}
		}
	}

	return maxStat
}

// GetSortedYears returns years in ascending order
func (s *Stats) GetSortedYears() []int {
	years := make([]int, 0, len(s.ByYear))
	for year := range s.ByYear {
		years = append(years, year)
	}
	sort.Ints(years)
	return years
}

// UserStat represents user with message count
type UserStat struct {
	Name  string
	Count int
}

// GetSortedUsers returns users sorted by message count (descending)
func GetSortedUsers(messagesByUser map[string]int) []UserStat {
	users := make([]UserStat, 0, len(messagesByUser))
	for name, count := range messagesByUser {
		users = append(users, UserStat{Name: name, Count: count})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Count > users[j].Count
	})
	return users
}

// GetMainUsers returns users with significant message count (>1% or top 10)
func GetMainUsers(messagesByUser map[string]int, total int) []UserStat {
	users := GetSortedUsers(messagesByUser)

	// Filter to main users (at least 1% of messages or top 10)
	threshold := total / 100
	if threshold < 10 {
		threshold = 10
	}

	result := make([]UserStat, 0)
	for i, u := range users {
		if u.Count >= threshold || i < 10 {
			result = append(result, u)
		}
		if i >= 10 && u.Count < threshold {
			break
		}
	}
	return result
}
