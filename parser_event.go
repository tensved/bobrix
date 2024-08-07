package bobrix

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"regexp"
	"slices"
)

func DefaultContractParser() func(evt *event.Event) *ServiceRequest {
	reMsg := regexp.MustCompile(
		`(@(?P<bot>\w+)\s+)*(-service:(?P<service>\w+)\s+)*(-method:(?P<method>\w+)\s)*(?P<inputs>.*)`,
	)

	reInputs := regexp.MustCompile(`-(\w+):"((?:\\"|[^"])*)"`)

	return func(evt *event.Event) *ServiceRequest {

		if evt.Type != event.EventMessage {
			return nil
		}

		eventMsg := evt.Content.AsMessage().Body

		match := reMsg.FindStringSubmatch(eventMsg)

		if len(match) == 0 {
			return nil
		}

		groups := make(map[string]string)
		for i, name := range reMsg.SubexpNames() {
			if i != 0 && name != "" && match[i] != "" {
				groups[name] = match[i]
			}
		}

		inputs := reInputs.FindAllStringSubmatch(groups["inputs"], -1)

		inputsData := make(map[string]any)

		for _, input := range inputs {
			if len(input) == 3 && input[1] != "" {
				if input[2] == "-" {
					inputsData[input[1]] = nil
					continue
				}

				inputsData[input[1]] = input[2]
			}
		}

		return &ServiceRequest{
			ServiceName: groups["service"],
			MethodName:  groups["method"],
			InputParams: inputsData,
		}
	}
}

type AudioMessageParserOpts struct {
	Downloader  Downloader
	ServiceName string
	MethodName  string
	InputName   string
}

// AudioMessageContractParser - audio message contract parser
// It is required to specify strictly the name of the service and method, as well as the name for Input,
// because when sending a voice message there is no possibility to specify parameters by text
func AudioMessageContractParser(opts *AudioMessageParserOpts) func(evt *event.Event) *ServiceRequest {
	if opts == nil {
		return nil
	}

	return func(evt *event.Event) *ServiceRequest {
		if evt.Type != event.EventMessage {
			return nil
		}

		if evt.Content.AsMessage().MsgType != event.MsgAudio {
			return nil
		}

		audioData, err := handleAudioMessage(opts.Downloader, evt)
		if err != nil {
			slog.Error("failed to handle audio message", "error", err)
			return nil
		}

		inputData := make(map[string]any, 1)

		inputData[opts.InputName] = audioData

		return &ServiceRequest{
			ServiceName: opts.ServiceName,
			MethodName:  opts.MethodName,
			InputParams: inputData,
		}
	}
}

type Downloader interface {
	Download(ctx context.Context, uri id.ContentURI) ([]byte, error)
}

func handleAudioMessage(bot Downloader, evt *event.Event) (string, error) {
	ctx := context.Background()

	var audioData string

	allowedMimeTypes := []string{
		"audio/webm",
		"audio/ogg",
		"audio/mpeg",
	}

	info := evt.Content.Raw["info"].(map[string]interface{})
	mimeType := info["mimetype"].(string)
	if slices.Contains(allowedMimeTypes, mimeType) {
		mxcURI, err := id.ContentURIString(evt.Content.Raw["url"].(string)).Parse()
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrParseMXCURI, err)
		}

		data, err := bot.Download(ctx, mxcURI)
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrDownloadFile, err)
		}

		encoded := base64.StdEncoding.EncodeToString(data)

		audioData = encoded
	} else {
		return "", fmt.Errorf("%w: %s", ErrInappropriateMimeType, mimeType)
	}

	return audioData, nil
}

type AutoParserOpts struct {
	ServiceName string
	MethodName  string
	InputName   string
}

// AutoRequestParser - auto request parser
// It is convenient to use it in cases when you need to handle situations
// when a user sends a message that should be sent immediately by a request
func AutoRequestParser(opts *AutoParserOpts) func(evt *event.Event) *ServiceRequest {

	return func(evt *event.Event) *ServiceRequest {
		if evt.Type != event.EventMessage {
			return nil
		}

		msg := evt.Content.AsMessage().Body

		inputData := make(map[string]any, 1)
		inputData[opts.InputName] = msg

		return &ServiceRequest{
			ServiceName: opts.ServiceName,
			MethodName:  opts.MethodName,
			InputParams: inputData,
		}
	}
}
