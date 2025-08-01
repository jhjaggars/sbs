package cmdlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Logger defines the interface for command logging
type Logger interface {
	LogCommand(command string, args []string, caller string) CommandContext
	IsEnabled() bool
	GetLevel() string
}

// CommandContext represents an active command logging context
type CommandContext interface {
	LogCompletion(success bool, exitCode int, errorMsg string, duration time.Duration)
}

// Config holds the configuration for command logging
type Config struct {
	Enabled  bool      // Enable/disable logging
	Level    string    // Log level: "debug", "info", "error"
	FilePath string    // Optional log file path
	Output   io.Writer // Direct output writer (for testing)
}

// commandLogger implements the Logger interface
type commandLogger struct {
	config Config
	logger *log.Logger
	mutex  sync.Mutex
}

// commandContext implements the CommandContext interface
type commandContext struct {
	logger    *commandLogger
	command   string
	args      []string
	caller    string
	startTime time.Time
}

// LogLevel represents different logging levels
type LogLevel int

const (
	LevelError LogLevel = iota
	LevelInfo
	LevelDebug
)

// Global logger instance
var globalLogger Logger = &noOpLogger{}
var globalMutex sync.RWMutex

// noOpLogger is a no-op implementation for when logging is disabled
type noOpLogger struct{}

func (n *noOpLogger) LogCommand(command string, args []string, caller string) CommandContext {
	return &noOpContext{}
}

func (n *noOpLogger) IsEnabled() bool {
	return false
}

func (n *noOpLogger) GetLevel() string {
	return ""
}

// noOpContext is a no-op implementation for when logging is disabled
type noOpContext struct{}

func (n *noOpContext) LogCompletion(success bool, exitCode int, errorMsg string, duration time.Duration) {
	// No-op
}

// NewCommandLogger creates a new command logger with the given configuration
func NewCommandLogger(config Config) Logger {
	if !config.Enabled {
		return &noOpLogger{}
	}

	logger := &commandLogger{
		config: config,
	}

	// Set up the output writer
	var output io.Writer
	if config.Output != nil {
		output = config.Output
	} else if config.FilePath != "" {
		// Create log file and directories if needed
		if err := os.MkdirAll(filepath.Dir(config.FilePath), 0755); err != nil {
			// Fall back to stderr if file creation fails
			output = os.Stderr
		} else {
			file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				// Fall back to stderr if file opening fails
				output = os.Stderr
			} else {
				output = file
			}
		}
	} else {
		// Default to stderr
		output = os.Stderr
	}

	logger.logger = log.New(output, "", log.LstdFlags)
	return logger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	globalMutex.RLock()
	defer globalMutex.RUnlock()
	return globalLogger
}

// LogCommand logs the start of a command execution
func (cl *commandLogger) LogCommand(command string, args []string, caller string) CommandContext {
	if !cl.config.Enabled {
		return &noOpContext{}
	}

	// If no caller provided, try to get it from runtime
	if caller == "" {
		caller = cl.getCaller()
	}

	ctx := &commandContext{
		logger:    cl,
		command:   command,
		args:      make([]string, len(args)),
		caller:    caller,
		startTime: time.Now(),
	}
	copy(ctx.args, args)

	return ctx
}

// LogCompletion logs the completion of a command execution
func (cc *commandContext) LogCompletion(success bool, exitCode int, errorMsg string, duration time.Duration) {
	level := LevelInfo
	if !success {
		level = LevelError
	}

	if !cc.logger.shouldLog(level, success) {
		return
	}

	cc.logger.mutex.Lock()
	defer cc.logger.mutex.Unlock()

	// Build the log message
	var msgBuilder strings.Builder
	msgBuilder.WriteString("[COMMAND] ")
	msgBuilder.WriteString(cc.command)

	if len(cc.args) > 0 {
		msgBuilder.WriteString(" ")
		msgBuilder.WriteString(strings.Join(cc.args, " "))
	}

	if cc.caller != "" {
		msgBuilder.WriteString(fmt.Sprintf(" (from: %s)", cc.caller))
	}

	// Add execution details
	msgBuilder.WriteString(fmt.Sprintf(" duration=%s exit_code=%d", duration, exitCode))

	if !success && errorMsg != "" {
		msgBuilder.WriteString(fmt.Sprintf(" error=%q", errorMsg))
	}

	// Log the message
	cc.logger.logger.Println(msgBuilder.String())
}

// IsEnabled returns whether logging is enabled
func (cl *commandLogger) IsEnabled() bool {
	return cl.config.Enabled
}

// GetLevel returns the current log level
func (cl *commandLogger) GetLevel() string {
	return cl.config.Level
}

// shouldLog determines if a message should be logged based on level and success
func (cl *commandLogger) shouldLog(level LogLevel, success bool) bool {
	if !cl.config.Enabled {
		return false
	}

	configLevel := cl.parseLogLevel(cl.config.Level)

	// For error-level config, only log failures
	if configLevel == LevelError {
		return !success
	}

	// For info and debug levels, log based on the level hierarchy
	return configLevel >= level
}

// parseLogLevel converts string level to LogLevel
func (cl *commandLogger) parseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// getCaller gets the caller information from the runtime stack
func (cl *commandLogger) getCaller() string {
	// Skip 3 frames: getCaller -> LogCommand -> actual caller
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}

	// Extract just the filename from the full path
	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}

// Helper functions for global logger

// LogCommandGlobal logs a command using the global logger
func LogCommandGlobal(command string, args []string, caller string) CommandContext {
	return GetGlobalLogger().LogCommand(command, args, caller)
}

// IsGlobalLoggingEnabled returns whether global logging is enabled
func IsGlobalLoggingEnabled() bool {
	return GetGlobalLogger().IsEnabled()
}

// GetGlobalLogLevel returns the global log level
func GetGlobalLogLevel() string {
	return GetGlobalLogger().GetLevel()
}

// Utility functions for getting caller information

// GetCaller returns caller information for the calling function
func GetCaller() string {
	// Skip 2 frames: GetCaller -> actual caller
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	filename := filepath.Base(file)
	return fmt.Sprintf("%s:%d", filename, line)
}
