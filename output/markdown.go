package output

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"telegram_message_analyzer/analyzer"
)

// russianMonths maps month numbers to Russian month names
var russianMonths = map[time.Month]string{
	time.January:   "Январь",
	time.February:  "Февраль",
	time.March:     "Март",
	time.April:     "Апрель",
	time.May:       "Май",
	time.June:      "Июнь",
	time.July:      "Июль",
	time.August:    "Август",
	time.September: "Сентябрь",
	time.October:   "Октябрь",
	time.November:  "Ноябрь",
	time.December:  "Декабрь",
}

// GenerateReports creates markdown reports for each year
func GenerateReports(stats *analyzer.Stats, outputDir string) error {
	// Create output directory if not exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate report for each year
	for _, year := range stats.GetSortedYears() {
		yearStats := stats.ByYear[year]
		filename := filepath.Join(outputDir, fmt.Sprintf("%d_report.md", year))

		content := generateYearReport(stats.ChatName, stats.ChatType, yearStats)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write report for %d: %w", year, err)
		}

		fmt.Printf("Создан отчет: %s\n", filename)
	}

	// Generate overall report
	overallFilename := filepath.Join(outputDir, "overall_report.md")
	overallContent := generateOverallReport(stats)
	if err := os.WriteFile(overallFilename, []byte(overallContent), 0644); err != nil {
		return fmt.Errorf("failed to write overall report: %w", err)
	}
	fmt.Printf("Создан общий отчет: %s\n", overallFilename)

	return nil
}

// generateYearReport creates markdown content for a specific year
func generateYearReport(chatName, chatType string, stats *analyzer.YearStats) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Отчет по чату за %d год\n\n", stats.Year))

	// Metadata section
	sb.WriteString("## Метаданные чата\n\n")
	sb.WriteString(fmt.Sprintf("- **Название чата:** %s\n", chatName))
	sb.WriteString(fmt.Sprintf("- **Тип:** %s\n", chatType))
	sb.WriteString(fmt.Sprintf("- **Период:** %s — %s\n",
		stats.FirstMessage.Format("02.01.2006"),
		stats.LastMessage.Format("02.01.2006")))
	sb.WriteString(fmt.Sprintf("- **Всего сообщений:** %d\n", stats.TotalMessages))
	sb.WriteString(fmt.Sprintf("- **Ответов:** %d\n", stats.RepliesCount))
	sb.WriteString(fmt.Sprintf("- **Пересланных:** %d\n", stats.ForwardedCount))
	sb.WriteString(fmt.Sprintf("- **Средняя длина сообщения:** %.1f символов\n\n", stats.AvgMessageLength))

	// Messages by user
	sb.WriteString("## Сообщения по участникам\n\n")
	sb.WriteString("| Участник | Сообщений | Доля |\n")
	sb.WriteString("|----------|-----------|------|\n")

	// Sort users by message count
	type userStat struct {
		name  string
		count int
	}
	users := make([]userStat, 0, len(stats.MessagesByUser))
	for name, count := range stats.MessagesByUser {
		users = append(users, userStat{name, count})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].count > users[j].count
	})

	for _, u := range users {
		percentage := float64(u.count) / float64(stats.TotalMessages) * 100
		sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", u.name, u.count, percentage))
	}
	sb.WriteString("\n")

	// Top 20 words by user
	sb.WriteString("## Топ-20 популярных слов по участникам\n\n")

	// Get main users sorted by message count
	mainUsers := analyzer.GetMainUsers(stats.MessagesByUser, stats.TotalMessages)

	for _, user := range mainUsers {
		if topWords, ok := stats.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", user.Name))
			sb.WriteString("| # | Слово | Количество |\n")
			sb.WriteString("|---|-------|------------|\n")
			for i, wc := range topWords {
				sb.WriteString(fmt.Sprintf("| %d | %s | %d |\n", i+1, wc.Word, wc.Count))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Most active time window
	sb.WriteString("## Самый активный период\n\n")
	sb.WriteString(fmt.Sprintf("**%02d:00 — %02d:00** — %d сообщений\n\n",
		stats.MostActiveWindow.StartHour,
		stats.MostActiveWindow.EndHour,
		stats.MostActiveWindow.Count))

	// Hourly activity breakdown
	sb.WriteString("### Активность по часам\n\n")
	sb.WriteString("| Период | Сообщений |\n")
	sb.WriteString("|--------|----------|\n")
	windows := []struct{ start, end int }{
		{0, 2}, {2, 4}, {4, 6}, {6, 8}, {8, 10}, {10, 12},
		{12, 14}, {14, 16}, {16, 18}, {18, 20}, {20, 22}, {22, 24},
	}
	for _, w := range windows {
		count := 0
		for h := w.start; h < w.end; h++ {
			count += stats.HourlyActivity[h]
		}
		sb.WriteString(fmt.Sprintf("| %02d:00-%02d:00 | %d |\n", w.start, w.end, count))
	}
	sb.WriteString("\n")

	// Most active month
	sb.WriteString("## Самый активный месяц\n\n")
	if stats.MostActiveMonth.Count > 0 {
		sb.WriteString(fmt.Sprintf("**%s %d** — %d сообщений\n\n",
			russianMonths[stats.MostActiveMonth.Month],
			stats.MostActiveMonth.Year,
			stats.MostActiveMonth.Count))
	}

	// Monthly activity breakdown
	sb.WriteString("### Активность по месяцам\n\n")
	sb.WriteString("| Месяц | Сообщений |\n")
	sb.WriteString("|-------|----------|\n")

	// Sort months
	type monthEntry struct {
		key   string
		count int
	}
	months := make([]monthEntry, 0, len(stats.MonthlyActivity))
	for k, v := range stats.MonthlyActivity {
		months = append(months, monthEntry{k, v})
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i].key < months[j].key
	})

	for _, m := range months {
		t, _ := time.Parse("2006-01", m.key)
		sb.WriteString(fmt.Sprintf("| %s %d | %d |\n", russianMonths[t.Month()], t.Year(), m.count))
	}
	sb.WriteString("\n")

	return sb.String()
}

// generateOverallReport creates a summary report across all years
func generateOverallReport(stats *analyzer.Stats) string {
	var sb strings.Builder

	sb.WriteString("# Общий отчет по чату\n\n")

	// Metadata section
	sb.WriteString("## Метаданные чата\n\n")
	sb.WriteString(fmt.Sprintf("- **Название чата:** %s\n", stats.ChatName))
	sb.WriteString(fmt.Sprintf("- **Тип:** %s\n", stats.ChatType))
	sb.WriteString(fmt.Sprintf("- **Период:** %s — %s\n",
		stats.Overall.FirstMessage.Format("02.01.2006"),
		stats.Overall.LastMessage.Format("02.01.2006")))
	sb.WriteString(fmt.Sprintf("- **Всего сообщений:** %d\n", stats.Overall.TotalMessages))
	sb.WriteString(fmt.Sprintf("- **Всего ответов:** %d\n", stats.Overall.RepliesCount))
	sb.WriteString(fmt.Sprintf("- **Всего пересланных:** %d\n", stats.Overall.ForwardedCount))
	sb.WriteString(fmt.Sprintf("- **Средняя длина сообщения:** %.1f символов\n\n", stats.Overall.AvgMessageLength))

	// Yearly summary
	sb.WriteString("## Статистика по годам\n\n")
	sb.WriteString("| Год | Сообщений | Ответов | Пересланных |\n")
	sb.WriteString("|-----|-----------|---------|-------------|\n")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		sb.WriteString(fmt.Sprintf("| %d | %d | %d | %d |\n", year, ys.TotalMessages, ys.RepliesCount, ys.ForwardedCount))
	}
	sb.WriteString("\n")

	// Messages by user (overall)
	sb.WriteString("## Сообщения по участникам (всего)\n\n")
	sb.WriteString("| Участник | Сообщений | Доля |\n")
	sb.WriteString("|----------|-----------|------|\n")

	type userStat struct {
		name  string
		count int
	}
	users := make([]userStat, 0, len(stats.Overall.MessagesByUser))
	for name, count := range stats.Overall.MessagesByUser {
		users = append(users, userStat{name, count})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].count > users[j].count
	})

	for _, u := range users {
		percentage := float64(u.count) / float64(stats.Overall.TotalMessages) * 100
		sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", u.name, u.count, percentage))
	}
	sb.WriteString("\n")

	// Messages by user per year
	sb.WriteString("## Сообщения по участникам (по годам)\n\n")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		sb.WriteString(fmt.Sprintf("### %d год\n\n", year))
		sb.WriteString("| Участник | Сообщений | Доля |\n")
		sb.WriteString("|----------|-----------|------|\n")

		yearUsers := make([]userStat, 0, len(ys.MessagesByUser))
		for name, count := range ys.MessagesByUser {
			yearUsers = append(yearUsers, userStat{name, count})
		}
		sort.Slice(yearUsers, func(i, j int) bool {
			return yearUsers[i].count > yearUsers[j].count
		})

		for _, u := range yearUsers {
			percentage := float64(u.count) / float64(ys.TotalMessages) * 100
			sb.WriteString(fmt.Sprintf("| %s | %d | %.1f%% |\n", u.name, u.count, percentage))
		}
		sb.WriteString("\n")
	}

	// Top 20 words by user (overall)
	sb.WriteString("## Топ-20 популярных слов по участникам (всего)\n\n")

	mainUsers := analyzer.GetMainUsers(stats.Overall.MessagesByUser, stats.Overall.TotalMessages)

	for _, user := range mainUsers {
		if topWords, ok := stats.Overall.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
			sb.WriteString(fmt.Sprintf("### %s\n\n", user.Name))
			sb.WriteString("| # | Слово | Количество |\n")
			sb.WriteString("|---|-------|------------|\n")
			for i, wc := range topWords {
				sb.WriteString(fmt.Sprintf("| %d | %s | %d |\n", i+1, wc.Word, wc.Count))
			}
			sb.WriteString("\n")
		}
	}

	// Top words by user per year
	sb.WriteString("## Топ-20 слов по участникам (по годам)\n\n")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		sb.WriteString(fmt.Sprintf("### %d год\n\n", year))

		yearMainUsers := analyzer.GetMainUsers(ys.MessagesByUser, ys.TotalMessages)

		for _, user := range yearMainUsers {
			if topWords, ok := ys.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
				sb.WriteString(fmt.Sprintf("#### %s\n\n", user.Name))
				sb.WriteString("| # | Слово | Количество |\n")
				sb.WriteString("|---|-------|------------|\n")
				for i, wc := range topWords {
					sb.WriteString(fmt.Sprintf("| %d | %s | %d |\n", i+1, wc.Word, wc.Count))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Most active time window overall
	sb.WriteString("## Самый активный период (общий)\n\n")
	sb.WriteString(fmt.Sprintf("**%02d:00 — %02d:00** — %d сообщений\n\n",
		stats.Overall.MostActiveWindow.StartHour,
		stats.Overall.MostActiveWindow.EndHour,
		stats.Overall.MostActiveWindow.Count))

	// Most active window by year
	sb.WriteString("### Самый активный период по годам\n\n")
	sb.WriteString("| Год | Период | Сообщений |\n")
	sb.WriteString("|-----|--------|----------|\n")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		sb.WriteString(fmt.Sprintf("| %d | %02d:00-%02d:00 | %d |\n",
			year,
			ys.MostActiveWindow.StartHour,
			ys.MostActiveWindow.EndHour,
			ys.MostActiveWindow.Count))
	}
	sb.WriteString("\n")

	// Most active month overall
	sb.WriteString("## Самый активный месяц (общий)\n\n")
	if stats.Overall.MostActiveMonth.Count > 0 {
		sb.WriteString(fmt.Sprintf("**%s %d** — %d сообщений\n\n",
			russianMonths[stats.Overall.MostActiveMonth.Month],
			stats.Overall.MostActiveMonth.Year,
			stats.Overall.MostActiveMonth.Count))
	}

	// Most active month by year
	sb.WriteString("### Самый активный месяц по годам\n\n")
	sb.WriteString("| Год | Месяц | Сообщений |\n")
	sb.WriteString("|-----|-------|----------|\n")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		if ys.MostActiveMonth.Count > 0 {
			sb.WriteString(fmt.Sprintf("| %d | %s | %d |\n",
				year,
				russianMonths[ys.MostActiveMonth.Month],
				ys.MostActiveMonth.Count))
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

// PrintConsoleStats prints statistics to console
func PrintConsoleStats(stats *analyzer.Stats) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("АНАЛИЗ ЧАТА: %s\n", stats.ChatName)
	fmt.Println(strings.Repeat("=", 60))

	fmt.Printf("\nПериод: %s — %s\n",
		stats.Overall.FirstMessage.Format("02.01.2006"),
		stats.Overall.LastMessage.Format("02.01.2006"))
	fmt.Printf("Всего сообщений: %d\n", stats.Overall.TotalMessages)

	// Messages by user
	fmt.Println("\n--- Сообщения по участникам ---")
	for name, count := range stats.Overall.MessagesByUser {
		percentage := float64(count) / float64(stats.Overall.TotalMessages) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", name, count, percentage)
	}

	// Print stats for each year
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		fmt.Printf("\n%s %d ГОД %s\n", strings.Repeat("-", 20), year, strings.Repeat("-", 20))
		fmt.Printf("Сообщений: %d\n", ys.TotalMessages)

		fmt.Println("\nСообщения по участникам:")
		for name, count := range ys.MessagesByUser {
			percentage := float64(count) / float64(ys.TotalMessages) * 100
			fmt.Printf("  %s: %d (%.1f%%)\n", name, count, percentage)
		}

		fmt.Println("\nТоп-20 слов по участникам:")
		yearMainUsers := analyzer.GetMainUsers(ys.MessagesByUser, ys.TotalMessages)
		for _, user := range yearMainUsers {
			if topWords, ok := ys.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
				fmt.Printf("\n  [%s]:\n", user.Name)
				for i, wc := range topWords {
					if i >= 10 {
						break // Show only top 10 in console
					}
					fmt.Printf("    %2d. %s (%d)\n", i+1, wc.Word, wc.Count)
				}
			}
		}

		fmt.Printf("\nСамый активный период: %02d:00-%02d:00 (%d сообщений)\n",
			ys.MostActiveWindow.StartHour,
			ys.MostActiveWindow.EndHour,
			ys.MostActiveWindow.Count)

		if ys.MostActiveMonth.Count > 0 {
			fmt.Printf("Самый активный месяц: %s %d (%d сообщений)\n",
				russianMonths[ys.MostActiveMonth.Month],
				ys.MostActiveMonth.Year,
				ys.MostActiveMonth.Count)
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
}
