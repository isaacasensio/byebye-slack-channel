package slack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaveChannel_returns_error_when_channel_does_not_exist(t *testing.T) {
	server := httpServerForLeaveChannel(t)
	defer server.Close()

	serverURL := fmt.Sprintf("%s/", server.URL)
	client := NewSlackClient(serverURL, "a-token")

	err := client.LeaveChannel("unknown")

	assert.EqualError(t, err, "channel_not_found")
}

func TestLeaveChannel_returns_nil_when_user_leaves_the_channel(t *testing.T) {
	server := httpServerForLeaveChannel(t)
	defer server.Close()

	serverURL := fmt.Sprintf("%s/", server.URL)
	client := NewSlackClient(serverURL, "a-token")

	err := client.LeaveChannel("C123")
	assert.Nil(t, err)
}

func TestGetAllUserChannels(t *testing.T) {
	tests := []struct {
		name             string
		userID           string
		expectedChannels []SlackChannel
		expectedError    bool
	}{
		{
			name: "returns a list of channels when there are no pages",
			expectedChannels: []SlackChannel{
				{ID: "C01", Name: "channel1"},
			},
			userID: "NO-CURSOR-USER",
		},
		{
			name: "returns the complete list of channels when pages",
			expectedChannels: []SlackChannel{
				{ID: "C01", Name: "channel1"},
				{ID: "C02", Name: "channel2"},
				{ID: "C03", Name: "channel3"},
			},
			userID: "WITH-CURSOR-USER",
		},
		{
			name:             "returns an empty list when something goes wrong",
			expectedChannels: []SlackChannel{},
			userID:           "ERROR",
			expectedError:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httpServer(t)
			defer server.Close()

			serverURL := fmt.Sprintf("%s/", server.URL)
			client := NewSlackClient(serverURL, "a-token")

			channels, err := client.FetchUserChannels(test.userID)

			assert.Equal(t, test.expectedError, err != nil)
			assert.Equal(t, test.expectedChannels, channels)
		})
	}
}

func httpServerForLeaveChannel(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		assert.NoError(t, err)

		var responseBody string

		switch userID := values.Get("channel"); userID {
		case "C123":
			responseBody = `{"ok": true}`
		default:
			responseBody = `
					{
						"ok": false,
						"error": "channel_not_found"
					}		
				`
		}
		_, err = w.Write([]byte(responseBody))
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
}

func httpServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		values, err := url.ParseQuery(string(body))
		assert.NoError(t, err)

		var responseBody string
		statusCode := http.StatusOK

		switch userID := values.Get("user"); userID {
		case "NO-CURSOR-USER":
			responseBody = `{
				"ok": true, 
				"channels": [
					{"id": "C01","name": "channel1"}
				]
			}`
		case "WITH-CURSOR-USER":
			if values.Get("cursor") == "1" {
				responseBody = `{
					"ok": true, 
					"channels": [
						{"id": "C03","name": "channel3"}
					]
				}`
			} else {
				responseBody = `{
					"ok": true, 
					"channels": [
						{"id": "C01","name": "channel1"},
						{"id": "C02","name": "channel2"}
					],
					"response_metadata": {
     		   			"next_cursor": "1"
    				}
				}`
			}
		default:
			statusCode = http.StatusInternalServerError
		}
		_, err = w.Write([]byte(responseBody))
		assert.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
	}))
}
