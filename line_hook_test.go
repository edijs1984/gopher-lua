package lua

import (
	"context"
	"testing"
)

func containsHookLine(events []HookEvent, line int) bool {
	for _, event := range events {
		if event.Line == line {
			return true
		}
	}
	return false
}

func TestHostLineHookSourceTransitions(t *testing.T) {
	L := NewState()
	defer L.Close()

	var events []HookEvent
	L.SetLineHook(func(_ *LState, event HookEvent) {
		events = append(events, event)
	})

	if err := L.DoString("local x = 1\nx = x + 1\nreturn x"); err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("line hook emitted no events")
	}
	if events[0].Line != 1 || events[0].Depth != 0 {
		t.Fatalf("first hook event = %+v, want line 1 depth 0", events[0])
	}
	for _, line := range []int{1, 2, 3} {
		if !containsHookLine(events, line) {
			t.Fatalf("missing source line %d in %#v", line, events)
		}
	}
}

func TestHostLineHookNestedDepth(t *testing.T) {
	L := NewState()
	defer L.Close()

	var events []HookEvent
	L.SetLineHook(func(_ *LState, event HookEvent) {
		events = append(events, event)
	})

	source := "local function f()\n" +
		"  local y = 1\n" +
		"  return y\n" +
		"end\n" +
		"return f()"

	if err := L.DoString(source); err != nil {
		t.Fatal(err)
	}

	foundNested := false
	for _, event := range events {
		if event.Depth > 0 {
			foundNested = true
			break
		}
	}
	if !foundNested {
		t.Fatalf("no nested Lua hook event in %#v", events)
	}
}

func TestHostLineHookWithContext(t *testing.T) {
	L := NewState()
	defer L.Close()
	L.SetContext(context.Background())

	count := 0
	L.SetLineHook(func(_ *LState, event HookEvent) {
		if event.Line > 0 {
			count++
		}
	})

	if err := L.DoString("local x = 1\nreturn x"); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("context VM loop emitted no line hook events")
	}
}

func TestHostLineHookCanBeCleared(t *testing.T) {
	L := NewState()
	defer L.Close()

	count := 0
	L.SetLineHook(func(_ *LState, _ HookEvent) {
		count++
	})
	if err := L.DoString("return 1"); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("installed hook was not called")
	}

	L.SetLineHook(nil)
	before := count
	if err := L.DoString("return 2"); err != nil {
		t.Fatal(err)
	}
	if count != before {
		t.Fatalf("cleared hook called %d extra times", count-before)
	}
}
