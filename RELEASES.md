# Релизная страница

На этой странице собраны готовые сборки TelegramMessageAnalyzer для основных настольных платформ и краткие инструкции по запуску.

## Windows (x86_64)
- **Файл:** `TelegramMessageAnalyzer_windows_amd64.exe`
- **Запуск:**
  ```powershell
  .\TelegramMessageAnalyzer_windows_amd64.exe -data "C:\\Path\\To\\ChatExport_*" -output "C:\\Path\\To\\Reports"
  ```
- **Сборка локально (при необходимости):**
  ```bash
  GOOS=windows GOARCH=amd64 go build -o TelegramMessageAnalyzer_windows_amd64.exe
  ```

## macOS (x86_64/arm64)
- **Файл:** `TelegramMessageAnalyzer_darwin_universal`
- **Запуск:**
  ```bash
  ./TelegramMessageAnalyzer_darwin_universal -data="/Users/you/Downloads/ChatExport_*" -output="/Users/you/Documents/Reports"
  ```
- **Сборка локально (универсальный бинарник):**
  ```bash
  GOOS=darwin GOARCH=amd64 go build -o TelegramMessageAnalyzer_darwin_amd64
  GOOS=darwin GOARCH=arm64 go build -o TelegramMessageAnalyzer_darwin_arm64
  lipo -create -output TelegramMessageAnalyzer_darwin_universal TelegramMessageAnalyzer_darwin_amd64 TelegramMessageAnalyzer_darwin_arm64
  ```

## Linux (x86_64)
- **Файл:** `TelegramMessageAnalyzer_linux_amd64`
- **Запуск:**
  ```bash
  ./TelegramMessageAnalyzer_linux_amd64 -data="~/Downloads/ChatExport_*" -output="~/reports"
  ```
- **Сборка локально (если нужно обновить версию):**
  ```bash
  GOOS=linux GOARCH=amd64 go build -o TelegramMessageAnalyzer_linux_amd64
  ```

## Общие требования
- Убедитесь, что в директории с бинарником есть папка `stopwords` и файлы из `parser`/`analyzer`, если приложение использует их в рантайме.
- Параметры `-data` и `-output` обязательны: путь к распакованному экспорту Telegram и директории, куда сохраняются отчёты.
- При первом запуске на macOS может потребоваться разрешить выполнение файла через System Settings → Privacy & Security.

## Проверка контрольных сумм (опционально)
Для дополнительной проверки целостности можно вычислить SHA-256:
```bash
shasum -a 256 TelegramMessageAnalyzer_*
```
