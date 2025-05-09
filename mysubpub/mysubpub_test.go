package mysubpub

import (
	"context"
	"testing"
	"time"
)

func TestMultipleSubscribers(t *testing.T) {
	sp := NewSubPub()
	subject := "testSubject"
	subscriberCount := 5
	messageCount := 100000
	receivedMessages := make([]chan interface{}, subscriberCount)
	for i := 0; i < subscriberCount; i++ {
		receivedMessages[i] = make(chan interface{}, messageCount)
	}

	for i := 0; i < subscriberCount; i++ {
		idx := i
		_, err := sp.Subscribe(subject, func(msg interface{}) {
			receivedMessages[idx] <- msg
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}
	for i := 0; i < messageCount; i++ {
		msg := i
		err := sp.Publish(subject, msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}
	for i := 0; i < subscriberCount; i++ {
		for j := 0; j < messageCount; j++ {
			select {
			case msg := <-receivedMessages[i]:
				if msg != j {
					t.Errorf("Subscriber %d received wrong message: got %v, want %v", i, msg, j)
				}

			case <-time.After(time.Second):
				t.Errorf("Subscriber %d did not receive message %d", i, j)
			}
		}
	}
}

func TestFIFO(t *testing.T) {
	sp := NewSubPub()
	subject := "testSubject"
	messageCount := 100000
	receivedMessages := make(chan interface{}, messageCount)
	_, err := sp.Subscribe(subject, func(msg interface{}) {
		receivedMessages <- msg
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	for i := 0; i < messageCount; i++ {
		msg := i
		err = sp.Publish(subject, msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}

	}
	for i := 0; i < messageCount; i++ {
		msg := <-receivedMessages
		if msg != i {
			t.Errorf("Received wrong message: got %v, want %v", msg, i)
		}
	}
}
func TestUnsubscribe(t *testing.T) {
	sp := NewSubPub()
	subject := "testSubject"
	messageCount := 10
	receivedMessages := make(chan interface{}, messageCount)
	sub, err := sp.Subscribe(subject, func(msg interface{}) {
		receivedMessages <- msg
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	err = sp.Publish(subject, "test1")
	if err != nil {
		t.Fatalf("Failed to publish first message: %v", err)
	}
	select {
	case msg := <-receivedMessages:
		if msg != "test1" {
			t.Errorf("Received wrong message: got %v, want test1", msg)
		}
	case <-time.After(time.Second):
		t.Error("Did not receive first message")
	}
	sub.Unsubscribe()

	err = sp.Publish(subject, "test2")
	if err != nil {
		t.Fatalf("Failed to publish second message:  %v", err)
	}

	select {
	case msg := <-receivedMessages:
		t.Errorf("Received message after unsubscribe: %v", msg)
	case <-time.After(time.Second):
	}

}
func TestClose(t *testing.T) {
	sp := NewSubPub()
	subject := "testSubject"

	received := make(chan interface{}, 1)
	_, err := sp.Subscribe(subject, func(msg interface{}) {
		received <- msg
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	if err := sp.Publish(subject, "test1"); err != nil {
		t.Fatalf("Failed to publish first message: %v", err)
	}

	select {
	case msg := <-received:
		if msg != "test1" {
			t.Errorf("Received wrong message: got %v, want test1", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for first message")
	}

	if err := sp.Close(context.Background()); err != nil {
		t.Fatalf("failed to close subscriber: %v", err)
	}

	if err := sp.Publish(subject, "test2"); err == nil {
		t.Error("Expected error when publishing after close, got nil")
	}

	select {
	case msg := <-received:
		t.Errorf("Received unexpected message after close: %v", msg)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestCloseWithContext(t *testing.T) {
	sp := NewSubPub()
	subject := "testSubject"

	subscriberCount := 3
	receivedMessages := make([]chan interface{}, subscriberCount)
	for i := 0; i < subscriberCount; i++ {
		receivedMessages[i] = make(chan interface{}, 100)
		idx := i
		_, err := sp.Subscribe(subject, func(msg interface{}) {
			receivedMessages[idx] <- msg
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	for i := 0; i < 50; i++ {
		if err := sp.Publish(subject, i); err != nil {
			t.Fatalf("Failed to publish message %d: %v", i, err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sp.Close(ctx); err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	sp = NewSubPub()
	subject = "testSubject2"

	for i := 0; i < subscriberCount; i++ {
		receivedMessages[i] = make(chan interface{}, 100)
		idx := i
		_, err := sp.Subscribe(subject, func(msg interface{}) {
			receivedMessages[idx] <- msg
		})
		if err != nil {
			t.Fatalf("failed to subscribe: %v", err)
		}
	}

	for i := 0; i < 50; i++ {
		if err := sp.Publish(subject, i); err != nil {
			t.Fatalf("Failed to publish message %d: %v", i, err)
		}
	}

	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := sp.Close(ctx); err != nil {
		if err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
		}
	}

	sp = NewSubPub()
	subject = "testSubject3"

	received := make(chan interface{}, 1)
	var err error
	_, err = sp.Subscribe(subject, func(msg interface{}) {
		received <- msg
	})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	if err := sp.Publish(subject, "test"); err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	select {
	case msg := <-received:
		if msg != "test" {
			t.Errorf("Received wrong message: got %v, want test", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for message")
	}

	if err := sp.Close(context.Background()); err != nil {
		t.Errorf("Expected successful close, got error: %v", err)
	}
}
