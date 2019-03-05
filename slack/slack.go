package slack

import "github.com/nlopes/slack"

// SlackClient wraps slack client api
type SlackClient struct {
	api *slack.Client
}

type ChannelLeaver interface {
	LeaveChannel(channelID string) error
}

type ChannelRetriever interface {
	FetchUserChannels(userID string) ([]SlackChannel, error)
}

type ChannelLeaverRetriever interface {
	ChannelLeaver
	ChannelRetriever
}

// NewSlackClient returns an instance of a SlackClient
func NewSlackClient(url string, token string) SlackClient {
	slack.APIURL = url
	api := slack.New(token)

	return SlackClient{
		api,
	}
}

// SlackChannel represents a slack channel
type SlackChannel struct {
	Name string
	ID   string
}

// LeaveChannel leaves the provided channel for the userID associated with
// the token
func (sc SlackClient) LeaveChannel(channelID string) error {
	_, err := sc.api.LeaveChannel(channelID)
	if err != nil {
		return err
	}
	return nil
}

// FetchUserChannels returns the list of public channels associated with a userId
func (sc SlackClient) FetchUserChannels(userID string) ([]SlackChannel, error) {
	channelNames := make([]SlackChannel, 0)
	cursor := ""

	for {

		r := &slack.GetConversationsForUserParameters{
			UserID: userID,
			Cursor: cursor,
		}

		channels, cc, err := sc.api.GetConversationsForUser(r)

		if err != nil {
			return []SlackChannel{}, err
		}

		for _, channel := range channels {
			c := SlackChannel{
				Name: channel.Name,
				ID:   channel.ID,
			}
			channelNames = append(channelNames, c)
		}

		if len(cc) == 0 {
			break
		}

		cursor = cc
	}

	return channelNames, nil
}
