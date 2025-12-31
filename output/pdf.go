package output

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"telegram_message_analyzer/analyzer"

	"github.com/signintech/gopdf"
)

const (
	pageWidth    = 595.28 // A4 width in points
	pageHeight   = 841.89 // A4 height in points
	marginLeft   = 40
	marginRight  = 40
	marginTop    = 40
	marginBottom = 40
	lineHeight   = 16
	titleSize    = 18
	headerSize   = 14
	normalSize   = 10
)

// PDFGenerator handles PDF report generation
type PDFGenerator struct {
	pdf      *gopdf.GoPdf
	fontPath string
	y        float64
}

// findFont tries to find a TTF font with Cyrillic support
func findFont() string {
	// Common font paths on different systems
	fontPaths := []string{
		// macOS
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
		"/Library/Fonts/Arial Unicode.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		// Linux
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/TTF/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		// Windows
		"C:/Windows/Fonts/arial.ttf",
		"C:/Windows/Fonts/arialuni.ttf",
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GeneratePDFReports creates PDF reports for each year
func GeneratePDFReports(stats *analyzer.Stats, outputDir string) error {
	fontPath := findFont()
	if fontPath == "" {
		return fmt.Errorf("не найден TTF шрифт с поддержкой кириллицы")
	}

	// Create output directory if not exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate report for each year
	for _, year := range stats.GetSortedYears() {
		yearStats := stats.ByYear[year]
		filename := filepath.Join(outputDir, fmt.Sprintf("%d_report.pdf", year))

		gen := &PDFGenerator{fontPath: fontPath}
		if err := gen.generateYearPDF(stats.ChatName, stats.ChatType, yearStats, filename); err != nil {
			return fmt.Errorf("failed to generate PDF for %d: %w", year, err)
		}

		fmt.Printf("Создан PDF отчет: %s\n", filename)
	}

	// Generate overall report
	overallFilename := filepath.Join(outputDir, "overall_report.pdf")
	gen := &PDFGenerator{fontPath: fontPath}
	if err := gen.generateOverallPDF(stats, overallFilename); err != nil {
		return fmt.Errorf("failed to generate overall PDF: %w", err)
	}
	fmt.Printf("Создан общий PDF отчет: %s\n", overallFilename)

	return nil
}

func (g *PDFGenerator) initPDF() error {
	g.pdf = &gopdf.GoPdf{}
	g.pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	err := g.pdf.AddTTFFont("font", g.fontPath)
	if err != nil {
		return fmt.Errorf("failed to add font: %w", err)
	}

	g.addPage()
	return nil
}

func (g *PDFGenerator) addPage() {
	g.pdf.AddPage()
	g.y = marginTop
}

func (g *PDFGenerator) checkPageBreak(neededHeight float64) {
	if g.y+neededHeight > pageHeight-marginBottom {
		g.addPage()
	}
}

func (g *PDFGenerator) writeTitle(text string) {
	g.checkPageBreak(30)
	g.pdf.SetFont("font", "", titleSize)
	g.pdf.SetX(marginLeft)
	g.pdf.SetY(g.y)
	g.pdf.Cell(nil, text)
	g.y += titleSize + 10
}

func (g *PDFGenerator) writeHeader(text string) {
	g.checkPageBreak(25)
	g.pdf.SetFont("font", "", headerSize)
	g.pdf.SetX(marginLeft)
	g.pdf.SetY(g.y)
	g.pdf.Cell(nil, text)
	g.y += headerSize + 8
}

func (g *PDFGenerator) writeSubHeader(text string) {
	g.checkPageBreak(20)
	g.pdf.SetFont("font", "", 12)
	g.pdf.SetX(marginLeft)
	g.pdf.SetY(g.y)
	g.pdf.Cell(nil, text)
	g.y += 12 + 6
}

func (g *PDFGenerator) writeLine(text string) {
	g.checkPageBreak(lineHeight)
	g.pdf.SetFont("font", "", normalSize)
	g.pdf.SetX(marginLeft)
	g.pdf.SetY(g.y)
	g.pdf.Cell(nil, text)
	g.y += lineHeight
}

func (g *PDFGenerator) writeTableRow(cols []string, widths []float64) {
	g.checkPageBreak(lineHeight)
	g.pdf.SetFont("font", "", normalSize)
	x := float64(marginLeft)
	for i, col := range cols {
		g.pdf.SetX(x)
		g.pdf.SetY(g.y)
		g.pdf.Cell(nil, col)
		if i < len(widths) {
			x += widths[i]
		}
	}
	g.y += lineHeight
}

func (g *PDFGenerator) addSpace(height float64) {
	g.y += height
}

func (g *PDFGenerator) generateYearPDF(chatName, chatType string, stats *analyzer.YearStats, filename string) error {
	if err := g.initPDF(); err != nil {
		return err
	}

	// Title
	g.writeTitle(fmt.Sprintf("Отчет по чату за %d год", stats.Year))
	g.addSpace(10)

	// Metadata
	g.writeHeader("Метаданные чата")
	g.writeLine(fmt.Sprintf("Название чата: %s", chatName))
	g.writeLine(fmt.Sprintf("Тип: %s", chatType))
	g.writeLine(fmt.Sprintf("Период: %s — %s",
		stats.FirstMessage.Format("02.01.2006"),
		stats.LastMessage.Format("02.01.2006")))
	g.writeLine(fmt.Sprintf("Всего сообщений: %d", stats.TotalMessages))
	g.writeLine(fmt.Sprintf("Ответов: %d", stats.RepliesCount))
	g.writeLine(fmt.Sprintf("Пересланных: %d", stats.ForwardedCount))
	g.writeLine(fmt.Sprintf("Средняя длина сообщения: %.1f символов", stats.AvgMessageLength))
	g.addSpace(10)

	// Messages by user
	g.writeHeader("Сообщения по участникам")
	colWidths := []float64{200, 80, 60}
	g.writeTableRow([]string{"Участник", "Сообщений", "Доля"}, colWidths)
	g.writeLine("─────────────────────────────────────────────────────")

	users := analyzer.GetSortedUsers(stats.MessagesByUser)
	for _, u := range users {
		if u.Count > 0 {
			percentage := float64(u.Count) / float64(stats.TotalMessages) * 100
			name := u.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			g.writeTableRow([]string{
				name,
				fmt.Sprintf("%d", u.Count),
				fmt.Sprintf("%.1f%%", percentage),
			}, colWidths)
		}
	}
	g.addSpace(10)

	// Top words by user
	g.writeHeader("Топ-20 слов по участникам")
	mainUsers := analyzer.GetMainUsers(stats.MessagesByUser, stats.TotalMessages)

	for _, user := range mainUsers {
		if topWords, ok := stats.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
			g.writeSubHeader(user.Name)
			wordWidths := []float64{30, 150, 80}
			for i, wc := range topWords {
				g.writeTableRow([]string{
					fmt.Sprintf("%d.", i+1),
					wc.Word,
					fmt.Sprintf("%d", wc.Count),
				}, wordWidths)
			}
			g.addSpace(5)
		}
	}

	// Time activity
	g.writeHeader("Активность по времени")
	g.writeLine(fmt.Sprintf("Самый активный период: %02d:00-%02d:00 (%d сообщений)",
		stats.MostActiveWindow.StartHour,
		stats.MostActiveWindow.EndHour,
		stats.MostActiveWindow.Count))

	if stats.MostActiveMonth.Count > 0 {
		g.writeLine(fmt.Sprintf("Самый активный месяц: %s %d (%d сообщений)",
			russianMonths[stats.MostActiveMonth.Month],
			stats.MostActiveMonth.Year,
			stats.MostActiveMonth.Count))
	}

	return g.pdf.WritePdf(filename)
}

func (g *PDFGenerator) generateOverallPDF(stats *analyzer.Stats, filename string) error {
	if err := g.initPDF(); err != nil {
		return err
	}

	// Title
	g.writeTitle("Общий отчет по чату")
	g.addSpace(10)

	// Metadata
	g.writeHeader("Метаданные чата")
	g.writeLine(fmt.Sprintf("Название чата: %s", stats.ChatName))
	g.writeLine(fmt.Sprintf("Тип: %s", stats.ChatType))
	g.writeLine(fmt.Sprintf("Период: %s — %s",
		stats.Overall.FirstMessage.Format("02.01.2006"),
		stats.Overall.LastMessage.Format("02.01.2006")))
	g.writeLine(fmt.Sprintf("Всего сообщений: %d", stats.Overall.TotalMessages))
	g.writeLine(fmt.Sprintf("Всего ответов: %d", stats.Overall.RepliesCount))
	g.writeLine(fmt.Sprintf("Всего пересланных: %d", stats.Overall.ForwardedCount))
	g.writeLine(fmt.Sprintf("Средняя длина сообщения: %.1f символов", stats.Overall.AvgMessageLength))
	g.addSpace(10)

	// Yearly summary
	g.writeHeader("Статистика по годам")
	yearWidths := []float64{60, 100, 80, 100}
	g.writeTableRow([]string{"Год", "Сообщений", "Ответов", "Пересланных"}, yearWidths)
	g.writeLine("─────────────────────────────────────────────────────")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		g.writeTableRow([]string{
			fmt.Sprintf("%d", year),
			fmt.Sprintf("%d", ys.TotalMessages),
			fmt.Sprintf("%d", ys.RepliesCount),
			fmt.Sprintf("%d", ys.ForwardedCount),
		}, yearWidths)
	}
	g.addSpace(10)

	// Messages by user (overall)
	g.writeHeader("Сообщения по участникам (всего)")
	colWidths := []float64{200, 80, 60}
	g.writeTableRow([]string{"Участник", "Сообщений", "Доля"}, colWidths)
	g.writeLine("─────────────────────────────────────────────────────")

	users := analyzer.GetSortedUsers(stats.Overall.MessagesByUser)
	for _, u := range users {
		percentage := float64(u.Count) / float64(stats.Overall.TotalMessages) * 100
		if percentage >= 0.1 { // Only show users with at least 0.1%
			name := u.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			g.writeTableRow([]string{
				name,
				fmt.Sprintf("%d", u.Count),
				fmt.Sprintf("%.1f%%", percentage),
			}, colWidths)
		}
	}
	g.addSpace(10)

	// Top words by user (overall)
	g.writeHeader("Топ-20 слов по участникам (всего)")
	mainUsers := analyzer.GetMainUsers(stats.Overall.MessagesByUser, stats.Overall.TotalMessages)

	for _, user := range mainUsers {
		if topWords, ok := stats.Overall.TopWordsByUser[user.Name]; ok && len(topWords) > 0 {
			g.writeSubHeader(user.Name)
			wordWidths := []float64{30, 150, 80}
			for i, wc := range topWords {
				g.writeTableRow([]string{
					fmt.Sprintf("%d.", i+1),
					wc.Word,
					fmt.Sprintf("%d", wc.Count),
				}, wordWidths)
			}
			g.addSpace(5)
		}
	}

	// Time activity
	g.writeHeader("Активность по времени")
	g.writeLine(fmt.Sprintf("Самый активный период: %02d:00-%02d:00 (%d сообщений)",
		stats.Overall.MostActiveWindow.StartHour,
		stats.Overall.MostActiveWindow.EndHour,
		stats.Overall.MostActiveWindow.Count))

	if stats.Overall.MostActiveMonth.Count > 0 {
		g.writeLine(fmt.Sprintf("Самый активный месяц: %s %d (%d сообщений)",
			russianMonths[stats.Overall.MostActiveMonth.Month],
			stats.Overall.MostActiveMonth.Year,
			stats.Overall.MostActiveMonth.Count))
	}

	g.addSpace(10)

	// Activity by year
	g.writeHeader("Самый активный период по годам")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		g.writeLine(fmt.Sprintf("%d: %02d:00-%02d:00 (%d сообщений)",
			year,
			ys.MostActiveWindow.StartHour,
			ys.MostActiveWindow.EndHour,
			ys.MostActiveWindow.Count))
	}

	g.addSpace(10)
	g.writeHeader("Самый активный месяц по годам")
	for _, year := range stats.GetSortedYears() {
		ys := stats.ByYear[year]
		if ys.MostActiveMonth.Count > 0 {
			g.writeLine(fmt.Sprintf("%d: %s (%d сообщений)",
				year,
				russianMonths[ys.MostActiveMonth.Month],
				ys.MostActiveMonth.Count))
		}
	}

	return g.pdf.WritePdf(filename)
}

// Helper to sort months
type monthEntry struct {
	key   string
	count int
}

func sortMonths(monthly map[string]int) []monthEntry {
	months := make([]monthEntry, 0, len(monthly))
	for k, v := range monthly {
		months = append(months, monthEntry{k, v})
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i].key < months[j].key
	})
	return months
}

func formatMonth(key string) string {
	t, err := time.Parse("2006-01", key)
	if err != nil {
		return key
	}
	return fmt.Sprintf("%s %d", russianMonths[t.Month()], t.Year())
}
