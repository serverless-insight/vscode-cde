package cde

import (
	"bytes"
	"testing"
)

func TestMessage(t *testing.T) {

}
func TestMessage_Raw(t *testing.T) {
	msg := Message{
		msgType:    MSG_INIT,
		channel:    1,
		fromServer: true,
		content:    []byte("content"),
	}
	expected := []byte{byte(msg.msgType), byte(msg.channel), 'c', 'o', 'n', 't', 'e', 'n', 't'}

	actual := msg.Raw()

	if !bytes.Equal(actual, expected) {
		t.Errorf("Raw() = %v, want %v", actual, expected)
	}
}
