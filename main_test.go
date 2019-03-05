package main

import (
	"bytes"
	"github.com/isaacasensio/byebye-slack-channel/slack"
	"github.com/pkg/errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterChannels(t *testing.T) {
	tests := []struct {
		name              string
		retrievedChannels []slack.SlackChannel
		toExclude         []string
		expectedChannels  []slack.SlackChannel
	}{
		{
			name: "returns empty when no channels available",
		},
		{
			name: "returns empty when only one channel and it is excluded",
			retrievedChannels: []slack.SlackChannel{
				{ID: "C01", Name: "channel1"},
			},
			toExclude: []string{"channel1"},
		},
		{
			name:      "returns empty when no current channels but some to exclude",
			toExclude: []string{"channel1"},
		},
		{
			name: "returns only those channels that are not excluded",
			retrievedChannels: []slack.SlackChannel{
				{ID: "C01", Name: "channel1"},
				{ID: "C02", Name: "channel2"},
			},
			toExclude: []string{"channel1"},
			expectedChannels: []slack.SlackChannel{
				{ID: "C02", Name: "channel2"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			channels := filter(test.retrievedChannels, test.toExclude)
			assert.Equal(t, channels, test.expectedChannels)
		})
	}
}

func TestGetExcludedChannelsTT(t *testing.T) {
	tests := []struct {
		name             string
		fileContent      string
		expectedChannels []string
	}{
		{
			name:             "returns empty list when file is empty",
			fileContent:      "",
			expectedChannels: nil,
		},
		{
			name:             "excludes empty lines",
			fileContent:      "channel1\n\n \nchannel2\n",
			expectedChannels: []string{"channel1", "channel2"},
		},
		{
			name:             "returns the list of excluded channels from a file",
			fileContent:      "channel1\nchannel2\nchannel3",
			expectedChannels: []string{"channel1", "channel2", "channel3"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.fileContent)
			channels := getExcludedChannels(reader)
			assert.Equal(t, test.expectedChannels, channels)
		})
	}
}

type FakeClient struct {
	fetchMethod func(userID string) ([]slack.SlackChannel, error)
	leaveMethod func(channelID string) error
}

func (f FakeClient) FetchUserChannels(userID string) ([]slack.SlackChannel, error) {
	return f.fetchMethod(userID)
}
func (f FakeClient) LeaveChannel(channelID string) error { return f.leaveMethod(channelID) }

func leaveChannelFailsFunc() func(string) error {
	return func(channelID string) error {
		return errors.New("Leave channels failed")
	}
}

func fetchChannelFailsFunc() func(string) ([]slack.SlackChannel, error) {
	return func(userID string) ([]slack.SlackChannel, error) {
		return nil, errors.New("Fetch user channels failed")
	}
}

func TestRun_returns_error_when_fetching_channel_call_fails(t *testing.T) {
	fake := struct{ FakeClient }{}

	fake.leaveMethod = func(channelID string) error { return nil }
	fake.fetchMethod = fetchChannelFailsFunc()

	err := run(fake, "1234", nil, false)
	assert.EqualError(t, err, "Fetch user channels failed")
}


func TestRun_returns_error_when_leave_channel_call_fails(t *testing.T) {
	fake := struct{ FakeClient }{}

	fake.leaveMethod = leaveChannelFailsFunc()
	fake.fetchMethod = func(userID string) ([]slack.SlackChannel, error) {
		return []slack.SlackChannel{
			{ID: "C01", Name: "channel1"},
		}, nil
	}

	err := run(fake, "1234", bytes.NewReader([]byte("")), false)
	assert.EqualError(t, err, "Leave channels failed")
}

func TestRun_should_not_call_leave_channel_when_dryRun_is_enabled(t *testing.T)  {
	fake := struct{ FakeClient }{}

	fake.leaveMethod = leaveChannelFailsFunc()
	fake.fetchMethod = func(userID string) ([]slack.SlackChannel, error) {
		return []slack.SlackChannel{
			{ID: "C01", Name: "channel1"},
		}, nil
	}

	err := run(fake, "1234", bytes.NewReader([]byte("")), true)
	assert.Nil(t, err)
}

func TestRun_should_leave_all_non_excluded_channels(t *testing.T)  {
	fake := struct{ FakeClient }{}
	var calls int

	fake.leaveMethod = func(channelID string) error {
		assert.True(t, channelID == "C01" || channelID == "C04")
		calls++
		return nil
	}
	fake.fetchMethod = func(userID string) ([]slack.SlackChannel, error) {
		return []slack.SlackChannel{
			{ID: "C01", Name: "channel1"},
			{ID: "C02", Name: "channel2"},
			{ID: "C03", Name: "channel3"},
			{ID: "C04", Name: "channel4"},
		}, nil
	}

	err := run(fake, "1234", bytes.NewReader([]byte("channel2\nchannel3")), false)
	assert.Nil(t, err)
	assert.Equal(t, 2, calls)
}