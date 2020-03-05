package messaging_test

import (
	"github.com/rh-messaging/qdr-shipshape/pkg/messaging"
	"testing"
)

const (
	message = "Hello_World"
)

func TestGenerateMessageContent(t *testing.T) {
	message_len := len(message)
	_test := func(t *testing.T, expected string, _len int) {
		received := messaging.GenerateMessageContent(message, _len)
		if received != expected {
			t.Error(
				"Expected", expected,
				"Got", received,
			)
		}
	}

	_test(t, message[0:message_len-1], message_len-1)
	_test(t, message, message_len)
	_test(t, message+"H", message_len+1)
	_test(t, "", 0)
	_test(t, "", -1)
}
