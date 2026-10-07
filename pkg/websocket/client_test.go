package websocket

import (
	"encoding/json"
	"testing"
)

func TestClientQueuesMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		message any
	}{
		{name: "plain message", message: "hello"},
		{name: "structured message", message: map[string]string{"cmd": "user_joined"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(nil, "session-1", 1)
			client.SendMessage(test.message)

			select {
			case got := <-client.send:
				data, err := json.Marshal(got)
				if err != nil {
					t.Fatal(err)
				}
				want, err := json.Marshal(test.message)
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != string(want) {
					t.Fatalf("message = %s, want %s", data, want)
				}
			default:
				t.Fatal("message was not queued")
			}
		})
	}
}

func TestClientQueuesResponse(t *testing.T) {
	t.Parallel()

	client := NewClient(nil, "session-1", 1)
	client.SendResponse("request-1", "auth", "ok")

	select {
	case message := <-client.send:
		data, err := json.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"rID":"request-1","cmd":"auth","status":"ok"}`
		if string(data) != want {
			t.Fatalf("response = %s, want %s", data, want)
		}
	default:
		t.Fatal("response was not queued")
	}
}

func TestClientSessionID(t *testing.T) {
	t.Parallel()

	client := NewClient(nil, "session-1", 1)
	if got := client.SessionID(); got != "session-1" {
		t.Fatalf("SessionID() = %q, want %q", got, "session-1")
	}
}
