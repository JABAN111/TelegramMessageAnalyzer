package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"telegram_message_analyzer/analyzer"
	"telegram_message_analyzer/output"
	"telegram_message_analyzer/parser"
)

func main() {
	// Parse command line arguments
	dataDir := flag.String("data", "path_to_tg", "Directory with exported Telegram HTML files")
	outputDir := flag.String("output", "path_to_reports", "Directory for output markdown reports")
	flag.Parse()

	// Get absolute paths
	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: не удалось получить путь к данным: %v\n", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: не удалось получить путь для отчетов: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║         АНАЛИЗАТОР TELEGRAM ЧАТОВ                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Printf("\nИсходные данные: %s\n", absDataDir)
	fmt.Printf("Папка отчетов: %s\n\n", absOutputDir)

	// Check if data directory exists
	if _, err := os.Stat(absDataDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Ошибка: директория с данными не найдена: %s\n", absDataDir)
		os.Exit(1)
	}

	// Step 1: Parse HTML files
	fmt.Println("📖 Парсинг HTML файлов...")
	result, err := parser.ParseAllFiles(absDataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга: %v\n", err)
		os.Exit(1)
	}

	if len(result.Messages) == 0 {
		fmt.Println("Предупреждение: не найдено текстовых сообщений")
		os.Exit(0)
	}

	// Step 2: Analyze data
	fmt.Println("\n📊 Анализ данных...")
	stats := analyzer.Analyze(result)

	// Step 3: Print console statistics
	output.PrintConsoleStats(stats)

	// Step 4: Generate markdown reports
	fmt.Println("\n📝 Генерация MD отчетов...")
	if err := output.GenerateReports(stats, absOutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка генерации MD отчетов: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Generate PDF reports
	pdfDir := filepath.Join(absOutputDir, "pdf-report")
	fmt.Println("\n📄 Генерация PDF отчетов...")
	if err := output.GeneratePDFReports(stats, pdfDir); err != nil {
		fmt.Fprintf(os.Stderr, "Предупреждение: не удалось создать PDF отчеты: %v\n", err)
		// Don't exit - PDF is optional
	}

	fmt.Println("\n✅ Анализ завершен!")
	fmt.Printf("📁 MD отчеты: %s\n", absOutputDir)
	fmt.Printf("📁 PDF отчеты: %s\n", pdfDir)
}
