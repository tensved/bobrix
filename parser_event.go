package bobrix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"slices"

	"github.com/tensved/bobrix/mxbot"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type ContractParser func(evt *event.Event) *ServiceRequest

var (
	regexpMessage = regexp.MustCompile(
		`(-service:(?P<service>\w+)\s+)*(-method:(?P<method>\w+)\s)*(?P<inputs>.*)`,
	)

	regexpInputs = regexp.MustCompile(`-(\w+):"((?:\\"|[^"])*)"`)
)

const BobrixPromptTag = "bobrix.prompt"

// BobrixContractParser - master contract parser
// It is used for parsing a request from the bot
// Request should be in the form of a JSON object with the following structure (ServiceRequest)
// And it should have a tag "bobrix.prompt" in event.Content
func BobrixContractParser(bot mxbot.Bot) ContractParser {
	filters := []mxbot.Filter{
		mxbot.FilterMessageText(),
		//mxbot.FilterTagMe(bot),
	}

	return func(evt *event.Event) *ServiceRequest {
		for _, filter := range filters {
			if !filter(evt) {

				slog.Info("filter failed", "event", evt)
				return nil
			}
		}

		if evt.Content.Raw == nil {
			slog.Error("raw message is nil", "event", evt)
			return nil
		}

		requestData, exists := evt.Content.Raw[BobrixPromptTag]
		if !exists {
			slog.Error("missing request data", "event", evt)
			return nil
		}

		requestMap, valid := requestData.(map[string]any)
		if !valid {
			slog.Error("request data is not a valid map", "event", evt)
			return nil
		}

		reqJSON, err := json.Marshal(requestMap)
		if err != nil {
			slog.Error("failed to marshal request data", "error", err)
			return nil
		}

		var serviceRequest ServiceRequest
		if err := json.Unmarshal(reqJSON, &serviceRequest); err != nil {
			slog.Error("failed to unmarshal request data", "error", err)
			return nil
		}

		return &serviceRequest
	}
}

// DefaultContractParser - default contract parser
// it parses message text and extracts bot name, service name and method name
// it used specific regular expression for parsing
// it returns nil if message doesn't match the pattern
// message pattern example: @{botname} -service:{servicename} -method:{methodname} -{input1}:"{inputvalue1}" -{input2}:"{inputvalue2}"
func DefaultContractParser(bot mxbot.Bot) ContractParser {
	filters := []mxbot.Filter{
		mxbot.FilterMessageText(),
		mxbot.FilterTagMe(bot),
	}

	return func(evt *event.Event) *ServiceRequest {

		for _, filter := range filters {
			if !filter(evt) {

				return nil
			}
		}

		eventMsg := evt.Content.AsMessage().Body

		match := regexpMessage.FindStringSubmatch(eventMsg)

		if len(match) == 0 {
			return nil
		}

		groups := make(map[string]string)
		for i, name := range regexpMessage.SubexpNames() {
			if i != 0 && name != "" && match[i] != "" {
				groups[name] = match[i]
			}
		}

		inputs := regexpInputs.FindAllStringSubmatch(groups["inputs"], -1)

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
func AudioMessageContractParser(opts *AudioMessageParserOpts) ContractParser {
	if opts == nil {
		return nil
	}

	filters := []mxbot.Filter{
		mxbot.FilterMessageAudio(),
	}

	return func(evt *event.Event) *ServiceRequest {

		for _, filter := range filters {
			if !filter(evt) {
				return nil
			}
		}

		audioData, err := downloadAudioMessage(opts.Downloader, evt)
		if err != nil {
			slog.Error("failed to download audio message", "error", err)
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

type AutoParserOpts struct {
	Bot         mxbot.Bot
	ServiceName string
	MethodName  string
	InputName   string
}

// AutoRequestParser - auto request parser
// It is convenient to use it in cases when you need to handle situations
// when a user sends a message that should be sent immediately by a request
func AutoRequestParser(opts *AutoParserOpts) ContractParser {

	if opts == nil {
		return nil
	}

	filters := []mxbot.Filter{
		mxbot.FilterMessageText(),
		mxbot.FilterTagMeOrPrivate(opts.Bot),
	}

	return func(evt *event.Event) *ServiceRequest {
		for _, filter := range filters {
			if !filter(evt) {
				return nil
			}
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

type ImageMessageParserOpts struct {
	Downloader  mxbot.BotMedia
	ServiceName string
	MethodName  string
	InputName   string
}

// ImageMessageContractParser - image message contract parser
// It is required to specify strictly the name of the service and method, as well as the name for Input
// because when sending an image there is no possibility to specify parameters by text
func ImageMessageContractParser(opts *ImageMessageParserOpts) func(evt *event.Event) *ServiceRequest {

	return func(evt *event.Event) *ServiceRequest {
		if evt.Type != event.EventMessage {
			return nil
		}

		if evt.Content.AsMessage().MsgType != event.MsgImage {
			return nil
		}

		imageData, err := downloadImageMessage(opts.Downloader, evt)
		if err != nil {
			slog.Error("failed to download image message", "error", err)
			return nil
		}

		inputData := make(map[string]any, 1)

		inputData[opts.InputName] = imageData

		return &ServiceRequest{
			ServiceName: opts.ServiceName,
			MethodName:  opts.MethodName,
			InputParams: inputData,
		}
	}
}

// downloadAudioMessage - download audio message
// It returns base64 encoded audio data
func downloadAudioMessage(bot Downloader, evt *event.Event) (string, error) {
	allowedMimeTypes := []string{
		"audio/webm",
		"audio/ogg",
		"audio/mpeg",
	}

	return downloadMediaMessage(bot, evt, allowedMimeTypes)
}

// downloadImageMessage - download image message
// It returns base64 encoded image data
func downloadImageMessage(bot Downloader, evt *event.Event) (string, error) {
	allowedMimeTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
	}

	return downloadMediaMessage(bot, evt, allowedMimeTypes)
}

// decryptMatrixFile decrypts a file encrypted using the Matrix protocol
func decryptMatrixFile(data []byte, fileInfo map[string]interface{}) ([]byte, error) {
	// Getting information about the encrypted file
	iv, ok := fileInfo["iv"].(string)
	if !ok {
		return nil, fmt.Errorf("iv not found in file info")
	}

	key, ok := fileInfo["key"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key not found in file info")
	}

	k, ok := key["k"].(string)
	if !ok {
		return nil, fmt.Errorf("key.k not found in file info")
	}

	// Checking the protocol version
	version, _ := fileInfo["v"].(string)
	if version != "v2" {
		return nil, fmt.Errorf("unsupported encryption version: %s", version)
	}

	// Checking the algorithm
	alg, _ := key["alg"].(string)
	if alg != "A256CTR" {
		return nil, fmt.Errorf("unsupported encryption algorithm: %s", alg)
	}

	// Decoding base64 with URL-safe format support
	ivBytes, err := base64.RawURLEncoding.DecodeString(iv)
	if err != nil {
		return nil, fmt.Errorf("failed to decode iv: %w", err)
	}

	keyBytes, err := base64.RawURLEncoding.DecodeString(k)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	// Create an AES cipher
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create a CTR mode
	stream := cipher.NewCTR(block, ivBytes)

	// Расшифровываем данные
	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)

	return decrypted, nil
}

// downloadMediaMessage - download media message
// It returns base64 encoded media data and checks if the mime type is allowed
func downloadMediaMessage(bot Downloader, evt *event.Event, allowedMimeTypes []string) (string, error) {
	ctx := context.Background()

	info := evt.Content.Raw["info"].(map[string]interface{})
	mimeType := info["mimetype"].(string)

	if !slices.Contains(allowedMimeTypes, mimeType) {
		return "", fmt.Errorf("%w: %s", ErrInappropriateMimeType, mimeType)
	}

	var url string
	var fileInfo map[string]interface{}

	if file, ok := evt.Content.Raw["file"].(map[string]interface{}); ok {
		// if evt was decrypted
		url, ok = file["url"].(string)
		if !ok {
			return "", fmt.Errorf("%w: url not found in file structure", ErrDownloadFile)
		}
		fileInfo = file
	} else {
		// if evt wasn't encrypted
		url, ok = evt.Content.Raw["url"].(string)
		if !ok {
			return "", fmt.Errorf("%w: url not found in message content", ErrDownloadFile)
		}
	}

	mxcURI, err := id.ContentURIString(url).Parse()
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrParseMXCURI, err)
	}

	data, err := bot.Download(ctx, mxcURI)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrDownloadFile, err)
	}

	// If the file is encrypted, decrypt it
	if fileInfo != nil {
		data, err = decryptMatrixFile(data, fileInfo)
		if err != nil {
			return "", fmt.Errorf("%w: failed to decrypt file: %v", ErrDownloadFile, err)
		}
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}
