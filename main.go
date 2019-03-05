package main

import (
	"bufio"
	"github.com/isaacasensio/byebye-slack-channel/slack"
	"io"
	"os"
	"strings"

	wrapped "github.com/nlopes/slack"
)

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/genuinetools/pkg/cli"
	"github.com/sirupsen/logrus"
)

func main() {
	var tokenPath string
	var excludeChannelsFilePath string
	var slackUserID string
	var dryRun bool

	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "byebye-slack-channel"
	p.Description = "A simple tool to leave slack channels in bulk."

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)

	p.FlagSet.StringVar(&tokenPath, "token-path", "", "path to a file that contains your Slack token")
	p.FlagSet.StringVar(&excludeChannelsFilePath, "exclude-channels-path", "", "path to a file that contains a list of channels to exclude")
	p.FlagSet.StringVar(&slackUserID, "user-id", "", "slack userID")
	p.FlagSet.BoolVar(&dryRun, "dry-run", false, "runs command without removing the channels")

	// Set the before function.
	p.Before = func(ctx context.Context) error {

		if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
			return errors.New("slack token file not found")
		}

		if _, err := os.Stat(excludeChannelsFilePath); os.IsNotExist(err) {
			return errors.New("exclude channels file not found")
		}

		if len(slackUserID) == 0 {
			return fmt.Errorf("invalid userID: %s", slackUserID)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {

		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		var cancel context.CancelFunc
		_, cancel = context.WithCancel(ctx)
		go func() {
			for sig := range c {
				cancel()
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		token, err := readSlackToken(tokenPath)
		if err != nil {
			return err
		}

		excludeFileReader, err := readFile(excludeChannelsFilePath)
		if err != nil {
			return err
		}

		client := slack.NewSlackClient(wrapped.APIURL, token)

		return run(client, slackUserID, excludeFileReader, dryRun)
	}

	// Run our program.
	p.Run()
}

func run(slackClient slack.ChannelLeaverRetriever, userID string, excludeFileReader io.Reader, dryRun bool) error {

	channels, err := slackClient.FetchUserChannels(userID)
	if err != nil {
		return err
	}

	excludedChannels := getExcludedChannels(excludeFileReader)
	channelsToRemove := filter(channels, excludedChannels)

	for _, channel := range channelsToRemove {
		if dryRun {
			logrus.Warn("Dry-run enabled. This command will NOT make a user leaves any chanel.")
		}
		logrus.Infof("leaving channel %s...", channel.Name)
		if !dryRun {
			err := slackClient.LeaveChannel(channel.ID)
			if err != nil {
				return err
			}
		}
	}

	logrus.Infof("Finished leaving slack channels for user %s", userID)
	return nil
}

func filter(current []slack.SlackChannel, toExclude []string) []slack.SlackChannel {
	var filtered []slack.SlackChannel
	for _, channel := range current {
		if !contains(toExclude, channel) {
			filtered = append(filtered, channel)
		}
	}
	return filtered
}

func contains(channels []string, channelToLookFor slack.SlackChannel) bool {
	for _, channel := range channels {
		if channel == channelToLookFor.Name {
			return true
		}
	}
	return false
}

func getExcludedChannels(content io.Reader) []string {

	var channels []string

	fileScanner := bufio.NewScanner(content)

	for fileScanner.Scan() {
		line := strings.TrimSpace(fileScanner.Text())
		if line != "" {
			channels = append(channels, line)
		}
	}

	return channels
}

func readFile(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return bufio.NewReader(file), nil
}
