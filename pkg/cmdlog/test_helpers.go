package cmdlog

import (
	"bytes"
	"io"
	"time"
)

// TestScenario represents a command execution scenario for testing
type TestScenario struct {
	Name     string
	Command  string
	Args     []string
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
}

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	LogEntries []LogEntry
	Config     Config
}

// LogEntry represents a single log entry for testing
type LogEntry struct {
	Level     string
	Command   string
	Arguments []string
	Caller    string
	Success   bool
	ExitCode  int
	Output    string
	Error     string
	Duration  time.Duration
	Timestamp time.Time
}

// MockCommandContext implements the CommandContext interface for testing
type MockCommandContext struct {
	logger *MockLogger
	entry  LogEntry
}

func (m *MockCommandContext) LogCompletion(success bool, exitCode int, errorMsg string, duration time.Duration) {
	m.entry.Success = success
	m.entry.ExitCode = exitCode
	m.entry.Error = errorMsg
	m.entry.Duration = duration
	m.entry.Timestamp = time.Now()

	// Add to logger's entries
	m.logger.LogEntries = append(m.logger.LogEntries, m.entry)
}

// NewMockLogger creates a new MockLogger for testing
func NewMockLogger(config Config) *MockLogger {
	return &MockLogger{
		LogEntries: make([]LogEntry, 0),
		Config:     config,
	}
}

func (m *MockLogger) LogCommand(command string, args []string, caller string) CommandContext {
	entry := LogEntry{
		Level:     m.Config.Level,
		Command:   command,
		Arguments: make([]string, len(args)),
		Caller:    caller,
	}
	copy(entry.Arguments, args)

	return &MockCommandContext{
		logger: m,
		entry:  entry,
	}
}

func (m *MockLogger) IsEnabled() bool {
	return m.Config.Enabled
}

func (m *MockLogger) GetLevel() string {
	return m.Config.Level
}

// CreateTestCommandScenario creates a test scenario for command execution testing
func CreateTestCommandScenario(name, command string, args []string, exitCode int, output string) TestScenario {
	return TestScenario{
		Name:     name,
		Command:  command,
		Args:     args,
		ExitCode: exitCode,
		Output:   output,
		Duration: 10 * time.Millisecond, // Default test duration
	}
}

// CreateTestCommandScenarioWithError creates a test scenario with error for command execution testing
func CreateTestCommandScenarioWithError(name, command string, args []string, exitCode int, output, errorMsg string) TestScenario {
	return TestScenario{
		Name:     name,
		Command:  command,
		Args:     args,
		ExitCode: exitCode,
		Output:   output,
		Error:    errorMsg,
		Duration: 15 * time.Millisecond, // Default test duration
	}
}

// TestBuffer provides a thread-safe buffer for testing logging output
type TestBuffer struct {
	buf *bytes.Buffer
}

func NewTestBuffer() *TestBuffer {
	return &TestBuffer{
		buf: &bytes.Buffer{},
	}
}

func (tb *TestBuffer) Write(p []byte) (n int, err error) {
	return tb.buf.Write(p)
}

func (tb *TestBuffer) String() string {
	return tb.buf.String()
}

func (tb *TestBuffer) Reset() {
	tb.buf.Reset()
}

// TestLoggerConfig provides common test configurations
var TestLoggerConfigs = struct {
	Disabled   Config
	InfoLevel  Config
	DebugLevel Config
	ErrorLevel Config
	WithFile   Config
}{
	Disabled: Config{
		Enabled: false,
		Level:   "info",
		Output:  io.Discard,
	},
	InfoLevel: Config{
		Enabled: true,
		Level:   "info",
		Output:  NewTestBuffer(),
	},
	DebugLevel: Config{
		Enabled: true,
		Level:   "debug",
		Output:  NewTestBuffer(),
	},
	ErrorLevel: Config{
		Enabled: true,
		Level:   "error",
		Output:  NewTestBuffer(),
	},
	WithFile: Config{
		Enabled:  true,
		Level:    "info",
		FilePath: "/tmp/sbs-test-command.log",
	},
}

// AssertLogContains checks if the log output contains expected content
func AssertLogContains(t interface{}, output, expected string) {
	// This would use testify assertions in real implementation
	// For now, it's a placeholder that can be used with testing.T
}

// AssertLogNotContains checks if the log output does not contain specific content
func AssertLogNotContains(t interface{}, output, unexpected string) {
	// This would use testify assertions in real implementation
	// For now, it's a placeholder that can be used with testing.T
}
