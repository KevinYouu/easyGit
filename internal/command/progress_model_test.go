package command

import (
	"testing"
)

func TestNewProgressModel(t *testing.T) {
	commands := []CommandInfo{
		{
			Command:     "echo",
			Args:        []string{"test1"},
			Description: "Test command 1",
			LoadingMsg:  "Loading 1",
			SuccessMsg:  "Success 1",
		},
		{
			Command:     "echo",
			Args:        []string{"test2"},
			Description: "Test command 2",
			LoadingMsg:  "Loading 2",
			SuccessMsg:  "Success 2",
		},
	}

	model := NewProgressModel(commands)

	if model == nil {
		t.Fatal("NewProgressModel returned nil")
	}

	if len(model.commands) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(model.commands))
	}

	if model.currentStep != 0 {
		t.Errorf("expected currentStep 0, got %d", model.currentStep)
	}

	if model.total != len(commands) {
		t.Errorf("expected total %d, got %d", len(commands), model.total)
	}
}

func TestNewProgressModelWithoutSpinner(t *testing.T) {
	commands := []CommandInfo{
		{
			Command:     "echo",
			Args:        []string{"test"},
			Description: "Test command",
		},
	}

	model := NewProgressModelWithoutSpinner(commands)

	if model == nil {
		t.Fatal("NewProgressModelWithoutSpinner returned nil")
	}

	if len(model.commands) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(model.commands))
	}
}

func TestCommandInfo(t *testing.T) {
	cmd := CommandInfo{
		Command:     "git",
		Args:        []string{"status"},
		Description: "Check git status",
		LoadingMsg:  "Checking status...",
		SuccessMsg:  "Status checked",
	}

	if cmd.Command != "git" {
		t.Errorf("expected command %q, got %q", "git", cmd.Command)
	}

	if len(cmd.Args) != 1 || cmd.Args[0] != "status" {
		t.Errorf("expected args [status], got %v", cmd.Args)
	}
}

