package random

import (
	"testing"
)

func TestExecuteRandomly(t *testing.T) {
	// 测试概率执行
	executed := false

	funcs := []FuncProbability{
		{
			Function:    func() { executed = true },
			Probability: 1.0,
		},
	}

	ExecuteRandomly(funcs)

	if !executed {
		t.Error("function with probability 1.0 should always execute")
	}
}

func TestExecuteRandomlyMultipleFunctions(t *testing.T) {
	// 测试多个函数的情况
	count1 := 0
	count2 := 0
	count3 := 0

	funcs := []FuncProbability{
		{
			Function:    func() { count1++ },
			Probability: 0.33,
		},
		{
			Function:    func() { count2++ },
			Probability: 0.33,
		},
		{
			Function:    func() { count3++ },
			Probability: 0.34,
		},
	}

	// 执行多次看是否有函数被调用
	iterations := 100
	for i := 0; i < iterations; i++ {
		ExecuteRandomly(funcs)
	}

	total := count1 + count2 + count3
	if total != iterations {
		t.Errorf("expected %d total executions, got %d", iterations, total)
	}

	// 至少每个函数应该被执行一次(概率很高)
	if count1 == 0 {
		t.Error("function 1 was never executed")
	}
	if count2 == 0 {
		t.Error("function 2 was never executed")
	}
	if count3 == 0 {
		t.Error("function 3 was never executed")
	}
}

func TestExecuteRandomlyEmptyList(t *testing.T) {
	// 测试空列表不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ExecuteRandomly with empty list should not panic: %v", r)
		}
	}()

	ExecuteRandomly([]FuncProbability{})
}

func TestExecuteRandomlyZeroProbability(t *testing.T) {
	executed1 := false
	executed2 := false

	funcs := []FuncProbability{
		{
			Function:    func() { executed1 = true },
			Probability: 0.0,
		},
		{
			Function:    func() { executed2 = true },
			Probability: 1.0,
		},
	}

	ExecuteRandomly(funcs)

	if executed1 {
		t.Error("function with probability 0.0 should not execute")
	}
	if !executed2 {
		t.Error("function with probability 1.0 should execute")
	}
}
