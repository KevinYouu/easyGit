package command

import (
	"strings"
	"testing"
	"time"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

func TestRunFuncWithSpinnerOptionsPanic(t *testing.T) {
	// 任务 panic 时应转为 error 返回，而非让主协程在 <-errChan 永久阻塞
	result := make(chan error, 1)
	go func() {
		result <- RunFuncWithSpinnerOptions("loading", "done", func() error {
			panic("boom")
		})
	}()

	select {
	case err := <-result:
		if err == nil {
			t.Fatal("expected error when task panics")
		}
		if !strings.Contains(err.Error(), i18n.T("ui.task_panicked")) {
			t.Errorf("expected task_panicked in error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Errorf("expected panic value in error, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("runWithSpinner blocked forever on task panic")
	}
}
