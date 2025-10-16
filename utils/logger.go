package utils

import (
	"fmt"
	"time"
)

const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorGray    = "\033[90m"

	Bold = "\033[1m"

	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconFile    = "📄"
	IconFolder  = "📁"
	IconBackup  = "💾"
	IconDelete  = "🗑️"
	IconWorker  = "🔧"
	IconStats   = "📊"
	IconWatch   = "👀"
)

type Logger struct {
	EnableColors bool
	ShowTime     bool
}

func NewLogger(colors, showTime bool) *Logger {
	return &Logger{
		EnableColors: colors,
		ShowTime:     showTime,
	}
}

func (l *Logger) colorize(color, text string) string {
	if !l.EnableColors {
		return text
	}
	return color + text + ColorReset
}

func (l *Logger) timestamp() string {
	if !l.ShowTime {
		return ""
	}
	return l.colorize(ColorGray, fmt.Sprintf("[%s] ", time.Now().Format("00:00:00")))
}

func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s\n",
		l.timestamp(),
		l.colorize(ColorRed, IconError),
		l.colorize(ColorRed, msg))
}

func (l *Logger) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s\n",
		l.timestamp(),
		l.colorize(ColorGreen, IconSuccess),
		l.colorize(ColorGreen, msg))
}

func (l *Logger) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s\n",
		l.timestamp(),
		l.colorize(ColorYellow, IconWarning),
		l.colorize(ColorYellow, msg))
}

func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s\n",
		l.timestamp(),
		l.colorize(ColorCyan, IconInfo),
		l.colorize(ColorCyan, msg))
}

func (l *Logger) FileCreated(filename string) {
	fmt.Printf("%s%s %s %s\n",
		l.timestamp(),
		l.colorize(ColorGreen, IconFile),
		l.colorize(ColorWhite, "New file:"),
		l.colorize(ColorCyan+Bold, filename))
}

func (l *Logger) FileModified(filename string) {
	fmt.Printf("%s%s %s %s\n",
		l.timestamp(),
		l.colorize(ColorBlue, IconFile),
		l.colorize(ColorWhite, "Modified"),
		l.colorize(ColorCyan+Bold, filename))
}

func (l *Logger) FileRenamed(filename string) {
	fmt.Printf("%s%s %s %s\n",
		l.timestamp(),
		l.colorize(ColorMagenta, IconFile),
		l.colorize(ColorWhite, "Renamed"),
		l.colorize(ColorCyan+Bold, filename))
}

func (l *Logger) FileDeleted(filename string) {
	fmt.Printf("%s%s %s %s\n",
		l.timestamp(),
		l.colorize(ColorRed, IconDelete),
		l.colorize(ColorWhite, "Deleted"),
		l.colorize(ColorGray, filename))
}

func (l *Logger) BackupCreated(filename, backupName string) {
	fmt.Printf("%s%s %s %s → %s\n",
		l.timestamp(),
		l.colorize(ColorGreen, IconBackup),
		l.colorize(ColorWhite, "Backup:"),
		l.colorize(ColorCyan, filename),
		l.colorize(ColorGray, backupName))
}

func (l *Logger) BackupSkipped(filename, reason string) {
	fmt.Printf("%s%s %s %s (%s)\n",
		l.timestamp(),
		l.colorize(ColorYellow, "⏭"),
		l.colorize(ColorWhite, "Skipped:"),
		l.colorize(ColorGray, filename),
		l.colorize(ColorYellow, reason))
}

func (l *Logger) WorkerStarted(id int, filename string) {
	fmt.Printf("%s%s %s %s\n",
		l.timestamp(),
		l.colorize(ColorMagenta, IconWorker),
		l.colorize(ColorWhite, fmt.Sprintf("Worker #%d →", id)),
		l.colorize(ColorCyan, filename))
}

func (l *Logger) Stats(tracked, queueLen, queueCap, workers int) {
	fmt.Printf("\n%s%s %s\n",
		l.timestamp(),
		l.colorize(ColorCyan, IconStats),
		l.colorize(ColorWhite+Bold, "Statistics"))

	fmt.Printf("	%s Tracked files: %s\n",
		l.colorize(ColorGray, "*"),
		l.colorize(ColorGreen+Bold, fmt.Sprintf("%d", tracked)))

	fmt.Printf("	%s Queue: %s\n",
		l.colorize(ColorGray, "*"),
		l.colorize(ColorYellow+Bold, fmt.Sprintf("%d/%d", queueLen, queueCap)))

	fmt.Printf("	%s Active workers: %s\n",
		l.colorize(ColorGray, "*"),
		l.colorize(ColorMagenta+Bold, fmt.Sprintf("%d", workers)))
}

func (l *Logger) Headder(source, backup string, versions, workers int) {
	fmt.Println(l.colorize(ColorCyan+Bold, "\n╔════════════════════════════════════════════╗"))
	fmt.Println(l.colorize(ColorCyan+Bold, "║   📂 File Watcher & Auto-Backup CLI      ║"))
	fmt.Println(l.colorize(ColorCyan+Bold, "╚════════════════════════════════════════════╝\n"))

	fmt.Printf("%s %s %s\n",
		l.colorize(ColorWhite, IconWatch+"  Monitoring:"),
		l.colorize(ColorGreen+Bold, source),
		l.colorize(ColorGray, "(recursive)"))

	fmt.Printf("%s %s\n",
		l.colorize(ColorWhite, IconBackup+"  Backup to:"),
		l.colorize(ColorGreen+Bold, backup))

	fmt.Printf("%s %s\n",
		l.colorize(ColorWhite, "📦  Versions:"),
		l.colorize(ColorYellow+Bold, fmt.Sprintf("%d", versions)))

	fmt.Printf("%s %s\n",
		l.colorize(ColorWhite, IconWorker+"  Workers:"),
		l.colorize(ColorMagenta+Bold, fmt.Sprintf("%d", workers)))

	fmt.Println(l.colorize(ColorGray, "\n"+"----------------------------------"))
	fmt.Println(l.colorize(ColorYellow, "Press Ctrl+C to stop watching and exit."))
	fmt.Println(l.colorize(ColorGray, "----------------------------------\n"))
}

func (l *Logger) Shutdown() {
	fmt.Println(l.colorize(ColorYellow+Bold, "\n\n👋 Closing application..."))
}

func (l *Logger) ShutdownComplete(duration time.Duration) {
	fmt.Printf("%s %s in %s\n",
		l.colorize(ColorGreen, IconSuccess),
		l.colorize(ColorGreen+Bold, "Application closed"),
		l.colorize(ColorCyan, duration.Round(time.Millisecond).String()))
}
