package logs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWarning(t *testing.T) {
	// 捕获 stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Warning("test warning message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test warning message") {
		t.Errorf("Warning() output should contain message, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Info("test info message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test info message") {
		t.Errorf("Info() output should contain message, got: %s", output)
	}
}

func TestSuccess(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Success("test success message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test success message") {
		t.Errorf("Success() output should contain message, got: %s", output)
	}
}

func TestError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Error("test error message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test error message") {
		t.Errorf("Error() output should contain message, got: %s", output)
	}
}
