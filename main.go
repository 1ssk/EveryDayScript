package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"
)

var logger *log.Logger

func main() {
    // Инициализируем логирование в файл
    logFile, err := os.OpenFile("daemon.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("Ошибка создания лог-файла: %v\n", err)
        os.Exit(1)
    }
    defer logFile.Close()
    logger = log.New(logFile, "DAEMON: ", log.Ldate|log.Ltime|log.Lshortfile)

    logger.Println("Демон запущен")

    // Бесконечный цикл для ежедневного выполнения
    for {
        performDailyCommit()

        // Ждём до следующего дня (00:00 следующего дня)
        now := time.Now()
        nextMidnight := now.Truncate(24 * time.Hour).Add(24 * time.Hour)
        sleepDuration := nextMidnight.Sub(now)

        logger.Printf("Ждём до следующего выполнения: %v\n", sleepDuration)
        time.Sleep(sleepDuration)
    }
}

func performDailyCommit() {
    date := time.Now().Format("2006-01-02")
    logFileName := "activity.log"
    commitMsg := fmt.Sprintf("Daily commit on %s", date)

    logger.Printf("Начало ежедневного коммита на %s", date)

    // Шаг 1: Добавляем строку в файл
    f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        logger.Printf("Ошибка открытия файла %s: %v", logFileName, err)
        return // Продолжаем цикл, не прерываем демон
    }
    defer func() {
        if cerr := f.Close(); cerr != nil {
            logger.Printf("Ошибка закрытия файла %s: %v", logFileName, cerr)
        }
    }()

    if _, err := f.WriteString(fmt.Sprintf("Daily activity on %s\n", date)); err != nil {
        logger.Printf("Ошибка записи в файл %s: %v", logFileName, err)
        return
    }

    logger.Printf("Строка добавлена в %s", logFileName)

    // Шаг 2: Git add
    if err := runGitCommand("add", logFileName); err != nil {
        logger.Printf("Ошибка git add %s: %v", logFileName, err)
        return
    }
    logger.Println("Git add выполнен")

    // Шаг 3: Git commit
    if err := runGitCommand("commit", "-m", commitMsg); err != nil {
        // Проверяем, если ничего не изменилось (exit code 1, но stdout пустой)
        if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
            logger.Println("Git commit: ничего не изменилось (возможно, commit уже был сегодня)")
        } else {
            logger.Printf("Ошибка git commit: %v", err)
        }
        return
    }
    logger.Println("Git commit выполнен")

    // Шаг 4: Git push
    if err := runGitCommand("push", "origin", "main"); err != nil {
        logger.Printf("Ошибка git push: %v", err)
        return
    }
    logger.Println("Git push выполнен")

    logger.Printf("Ежедневный коммит на %s завершён успешно", date)
}

// Вспомогательная функция для выполнения git-команд с обработкой ошибок
func runGitCommand(args ...string) error {
    cmd := exec.Command("git", args...)
    output, err := cmd.CombinedOutput() // Захватываем stdout и stderr
    if err != nil {
        return fmt.Errorf("команда 'git %v' failed: %v\nOutput: %s", args, err, string(output))
    }
    return nil
}
