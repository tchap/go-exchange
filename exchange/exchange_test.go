// Copyright (c) 2014 The go-exchange AUTHORS
//
// Use of this source code is governed by The MIT License
// that can be found in the LICENSE file.

package exchange

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"testing"
)

func TestExchange_PublishSubcribe(t *testing.T) {
	var (
		eventsSent     int32 = 100
		eventsReceived int32

		chatTopic   = Topic("chatroom.golang")
		chatMessage = "Golang is the best in this cruel world"
	)

	exchange := New()

	_, err := exchange.Subscribe(chatTopic, func(topic Topic, event Event) {
		if !bytes.HasPrefix(topic, chatTopic) {
			t.Errorf("%q is not a prefix of %q", topic, chatTopic)
			return
		}
		if e := event.(string); e != chatMessage {
			t.Errorf("Unexpected event body received, expected %q, got %q", chatMessage, e)
		}
		atomic.AddInt32(&eventsReceived, 1)
	})
	if err != nil {
		t.Fatal(err)
	}

	extendedChatTopic := append(chatTopic, []byte(".czech")...)

	for i := int32(0); i != 2*eventsSent; i++ {
		if i%2 == 0 {
			// This should cause the handler to be invoked since chatTopic is
			// a prefix of extendedChatTopic.
			if err := exchange.Publish(extendedChatTopic, chatMessage); err != nil {
				t.Fatal(err)
			}
		} else {
			// This should not cause the handler to be invoked.
			if err := exchange.Publish(chatTopic[1:], ""); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := exchange.Terminate(); err != nil {
		t.Fatal(err)
	}

	if eventsReceived != eventsSent {
		t.Errorf("Expected the handler to be invoked %v times, was %v",
			eventsSent, eventsReceived)
	}
}

func ExampleExchange_PublishSubscribe() {
	exchange := New()

	exchange.Subscribe(Topic("git"), func(topic Topic, event Event) {
		fmt.Printf("Event received:\n  topic: %q\n  body:  %v\n", topic, event)
	})

	exchange.Publish(Topic("git.push"), "b839dc656e3e78647c09453b33652b389e37c07a")

	exchange.Terminate()
	// Output:
	// Event received:
	//   topic: "git.push"
	//   body:  b839dc656e3e78647c09453b33652b389e37c07a
}
