package eventbus

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestBusBasics(t *testing.T) {
	bus, err := NewEventBus()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	eventType := logical.EventType("someType")

	err = bus.Send(ctx, eventType, logical.NewEvent())
	if err != ErrNotStarted {
		t.Errorf("Expected not started error but got: %v", err)
	}

	bus.Start()

	err = bus.Send(ctx, eventType, logical.NewEvent())
	if err != nil {
		t.Errorf("Expected no error sending: %v", err)
	}

	ch, err := bus.Subscribe(ctx, eventType)
	if err != nil {
		t.Fatal(err)
	}

	event := logical.NewEvent()

	err = bus.Send(ctx, eventType, event)
	if err != nil {
		t.Error(err)
	}

	timeout := time.After(1 * time.Second)
	select {
	case message := <-ch:
		if message.GetEid() != event.GetEid() {
			t.Errorf("Got unexpected message: %+v", message)
		}
	case <-timeout:
		t.Error("Timeout waiting for message")
	}
}

func TestBus2Subscriptions(t *testing.T) {
	bus, err := NewEventBus()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	eventType1 := logical.EventType("someType1")
	eventType2 := logical.EventType("someType2")
	bus.Start()

	ch1, err := bus.Subscribe(ctx, eventType1)
	if err != nil {
		t.Fatal(err)
	}

	ch2, err := bus.Subscribe(ctx, eventType2)
	if err != nil {
		t.Fatal(err)
	}

	event1 := logical.NewEvent()
	event2 := logical.NewEvent()

	err = bus.Send(ctx, eventType2, event2)
	if err != nil {
		t.Error(err)
	}
	err = bus.Send(ctx, eventType1, event1)
	if err != nil {
		t.Error(err)
	}

	timeout := time.After(1 * time.Second)
	select {
	case message := <-ch1:
		if message.GetEid() != event1.GetEid() {
			t.Errorf("Got unexpected message: %v", message)
		}
	case <-timeout:
		t.Error("Timeout waiting for event1")
	}
	select {
	case message := <-ch2:
		if message.GetEid() != event2.GetEid() {
			t.Errorf("Got unexpected message: %v", message)
		}
	case <-timeout:
		t.Error("Timeout waiting for event2")
	}
}
