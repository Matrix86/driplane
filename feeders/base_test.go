package feeders

import (
	"fmt"
	"testing"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
)

// stubFeeder embeds Base and adds the missing OnEvent method
// so it satisfies the Feeder interface for testing purposes.
type stubFeeder struct {
	Base
}

func (s *stubFeeder) OnEvent(e *data.Event) {}

func TestBaseSetAndGetName(t *testing.T) {
	b := &Base{}
	b.setName("myfeeder")
	if b.Name() != "myfeeder" {
		t.Errorf("expected name 'myfeeder', got '%s'", b.Name())
	}
}

func TestBaseSetAndGetRuleName(t *testing.T) {
	b := &Base{}
	b.setRuleName("rule1")
	if b.Rule() != "rule1" {
		t.Errorf("expected rule 'rule1', got '%s'", b.Rule())
	}
}

func TestBaseSetID(t *testing.T) {
	b := &Base{}
	b.setID(42)
	if b.id != 42 {
		t.Errorf("expected id 42, got %d", b.id)
	}
}

func TestBaseSetBus(t *testing.T) {
	b := &Base{}
	bus := EventBus.New()
	b.setBus(bus)
	if b.bus == nil {
		t.Errorf("expected bus to be set")
	}
}

func TestBaseGetIdentifier(t *testing.T) {
	b := &Base{}
	b.setName("test")
	b.setID(7)
	expected := "test:7"
	if b.GetIdentifier() != expected {
		t.Errorf("expected identifier '%s', got '%s'", expected, b.GetIdentifier())
	}
}

func TestBaseGetIdentifierZeroID(t *testing.T) {
	b := &Base{}
	b.setName("feeder")
	b.setID(0)
	expected := "feeder:0"
	if b.GetIdentifier() != expected {
		t.Errorf("expected identifier '%s', got '%s'", expected, b.GetIdentifier())
	}
}

func TestBaseIsRunningDefault(t *testing.T) {
	b := &Base{}
	if b.IsRunning() {
		t.Errorf("expected IsRunning to be false by default")
	}
}

func TestBaseIsRunningTrue(t *testing.T) {
	b := &Base{isRunning: true}
	if !b.IsRunning() {
		t.Errorf("expected IsRunning to be true")
	}
}

func TestBaseStartStop(t *testing.T) {
	b := &Base{}
	// Start and Stop on Base are no-ops; just verify they don't panic
	b.Start()
	b.Stop()
}

func TestBasePropagate(t *testing.T) {
	b := &Base{}
	bus := EventBus.New()
	b.setBus(bus)
	b.setName("propagate_test")
	b.setRuleName("test_rule")
	b.setID(1)

	received := make(chan *data.Message, 1)
	bus.Subscribe(b.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	msg := data.NewMessage("hello")
	b.Propagate(msg)

	got := <-received

	if got.GetMessage() != "hello" {
		t.Errorf("expected main message 'hello', got '%v'", got.GetMessage())
	}

	extra := got.GetExtra()
	if v, ok := extra["source_feeder"]; !ok || v != "propagate_test" {
		t.Errorf("expected source_feeder 'propagate_test', got '%v'", v)
	}
	if v, ok := extra["source_feeder_rule"]; !ok || v != "test_rule" {
		t.Errorf("expected source_feeder_rule 'test_rule', got '%v'", v)
	}
	if v, ok := extra["rule_name"]; !ok || v != "test_rule" {
		t.Errorf("expected rule_name 'test_rule', got '%v'", v)
	}
}

func TestBasePropagateWithExtra(t *testing.T) {
	b := &Base{}
	bus := EventBus.New()
	b.setBus(bus)
	b.setName("extra_test")
	b.setRuleName("rule2")
	b.setID(5)

	received := make(chan *data.Message, 1)
	bus.Subscribe(b.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	msg := data.NewMessageWithExtra("payload", map[string]interface{}{
		"custom_key": "custom_value",
	})
	b.Propagate(msg)

	got := <-received
	extra := got.GetExtra()

	if v, ok := extra["custom_key"]; !ok || v != "custom_value" {
		t.Errorf("expected custom_key 'custom_value', got '%v'", v)
	}
	// source_feeder extras should still be injected
	if v, ok := extra["source_feeder"]; !ok || v != "extra_test" {
		t.Errorf("expected source_feeder 'extra_test', got '%v'", v)
	}
}

func TestRegisterAndNewFeeder(t *testing.T) {
	called := false
	factory := func(conf map[string]string) (Feeder, error) {
		called = true
		return &stubFeeder{}, nil
	}

	// Use a unique name to avoid collision with real feeders
	feederFactories["_unittest_feeder"] = factory

	bus := EventBus.New()
	f, err := NewFeeder("rule", "_unittest_feeder", map[string]string{}, bus, 99)
	if err != nil {
		t.Fatalf("NewFeeder returned error: %s", err)
	}
	if !called {
		t.Errorf("factory function was not called")
	}
	if f == nil {
		t.Fatal("expected non-nil feeder")
	}
	if f.Name() != "_unittest_feeder" {
		t.Errorf("expected name '_unittest_feeder', got '%s'", f.Name())
	}
	if f.Rule() != "rule" {
		t.Errorf("expected rule 'rule', got '%s'", f.Rule())
	}
	expected := fmt.Sprintf("_unittest_feeder:%d", 99)
	if f.GetIdentifier() != expected {
		t.Errorf("expected identifier '%s', got '%s'", expected, f.GetIdentifier())
	}

	// cleanup
	delete(feederFactories, "_unittest_feeder")
}

func TestNewFeederNotFound(t *testing.T) {
	bus := EventBus.New()
	_, err := NewFeeder("rule", "nonexistent_feeder", map[string]string{}, bus, 1)
	if err == nil {
		t.Errorf("expected error for unknown feeder name")
	}
}

func TestNewFeederFactoryError(t *testing.T) {
	factory := func(conf map[string]string) (Feeder, error) {
		return nil, fmt.Errorf("factory error")
	}

	feederFactories["_errtest_feeder"] = factory
	defer delete(feederFactories, "_errtest_feeder")

	bus := EventBus.New()
	f, err := NewFeeder("rule", "_errtest_feeder", map[string]string{}, bus, 1)
	if err == nil {
		t.Errorf("expected error from factory")
	}
	if f != nil {
		t.Errorf("expected nil feeder when factory returns error")
	}
}

func TestNewFeederFactoryReturnsNil(t *testing.T) {
	factory := func(conf map[string]string) (Feeder, error) {
		return nil, nil
	}

	feederFactories["_niltest_feeder"] = factory
	defer delete(feederFactories, "_niltest_feeder")

	bus := EventBus.New()
	f, err := NewFeeder("rule", "_niltest_feeder", map[string]string{}, bus, 1)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if f != nil {
		t.Errorf("expected nil feeder when factory returns nil")
	}
}
