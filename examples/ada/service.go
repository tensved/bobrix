package ada

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/tensved/bobrix/contracts"
)

func NewADAService(adaHost string) *contracts.Service {

	return &contracts.Service{
		Name: "ada",
		Description: map[string]string{
			"en": "Ada is an AI language model trained by OpenAI.",
			"ru": "Ada — это языковая модель искусственного интеллекта, обученная OpenAI.",
		},
		Methods: map[string]*contracts.Method{
			"generate": {
				Name: "generate",
				Description: map[string]string{
					"en": "Generates text using Ada's language model.",
					"ru": "Генерирует текст с использованием языковой модели Ады.",
				},
				Inputs: []contracts.Input{
					{
						Name: "prompt",
						Description: map[string]string{
							"en": "The prompt to generate text from.",
							"ru": "Запрос на генерацию текста.",
						},
						Type: "text",
					},
					{
						Name:         "response_type",
						Type:         "text",
						DefaultValue: "text",
					},
					{
						Name: "audio",
						Type: "audio",
					},
				},
				Outputs: []contracts.Output{
					{
						Name: "text",
						Description: map[string]string{
							"en": "The generated text.",
							"ru": "Cгенерированный текст",
						},
					},
				},
				Handler: NewADAHandler(adaHost),
			},
		},
		Pinger: contracts.NewWSPinger(
			contracts.WSOptions{Host: adaHost, Path: "/", Schema: "wss"},
		),
	}
}

func NewADAHandler(adaHost string) *contracts.Handler {
	adaURL := url.URL{
		Scheme: "wss",
		Host:   adaHost,
		Path:   "/",
	}
	return &contracts.Handler{
		Name: "ada",
		Description: map[string]string{
			"en": "Generates text using Ada's language model.",
			"ru": "Генерирует текст, используя языковую модель Ада.",
		},
		Args: map[string]any{
			"URL": adaURL.String(),
		},
		Do: func(c contracts.HandlerContext) error {
			ctx := c.Context()

			slog.Debug("context history", "ctx", c.Messages())

			conn, _, err := websocket.DefaultDialer.DialContext(
				ctx,
				adaURL.String(),
				nil,
			)

			if err != nil {
				slog.Error("failed to dial", "err", err)
				return err
			}

			defer func(conn *websocket.Conn) {
				err := conn.Close()
				if err != nil {
					slog.Error("failed to close connection", "err", err)
				}
			}(conn)

			responseType := c.Get("response_type")
			if responseType == nil {
				return fmt.Errorf("response type not found")
			}

			responseTypeString := responseType.(string)

			message := InputMessage{
				ClientName:   "bot",
				Username:     "username",
				RequestType:  "text",
				ResponseType: responseTypeString,
			}

			audioData := c.Get("audio")
			if audioData != nil {
				message.RequestType = "speech"

				audio, ok := audioData.(string)
				if !ok {
					return fmt.Errorf("audio not found")
				}
				message.Speech = audio
			} else {

				textData := c.Get("prompt")
				if textData == nil {
					return fmt.Errorf("prompt not found")
				}

				message.RequestType = "text"
				message.Text = textData.(string)

			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				slog.Error("failed to marshal Input JSON", "err", err)
				return fmt.Errorf("failed to marshal Input JSON: %w", err)
			}

			slog.Info("sending message", "message", string(messageBytes))

			err = conn.WriteMessage(websocket.TextMessage, messageBytes)
			if err != nil {
				slog.Error("failed to send Input JSON", "err", err)
				return fmt.Errorf("failed to send Input JSON: %w", err)
			}

			_, p, err := conn.ReadMessage()
			if err != nil {
				slog.Error("failed to read OutputMessage JSON", "err", err)
				return fmt.Errorf("failed to read OutputMessage JSON: %w", err)
			}

			var responseMessage OutputMessage
			err = json.Unmarshal(p, &responseMessage)
			if err != nil {
				slog.Error("failed to unmarshal OutputMessage JSON", "err", err)
				return fmt.Errorf("failed to unmarshal OutputMessage JSON: %w", err)
			}

			slog.Debug("Received new message from ADA", "answer", responseMessage.Answer, "RESPONSE", responseMessage)

			return c.JSON(map[string]any{
				"text": responseMessage.Answer,
			})
		},
	}
}

type InputMessage struct {
	ClientName      string `json:"client"`            // "web", "tg", "telegram", "bot"
	Username        string `json:"username"`          // any
	RequestType     string `json:"request_type"`      // "text" or "speech"
	ResponseType    string `json:"response_type"`     // "text" or "speech"
	IsSearchEnabled bool   `json:"is_search_enabled"` // default true
	Speech          string `json:"speech"`            // only webm or ogg, base64
	Text            string `json:"text"`              // text message
}

type OutputMessage struct {
	Answer  string `json:"response"` // text if response_type=text, audio if response_type=speech
	Message string `json:"msg"`      // only if event="error"
	Event   string `json:"event"`    // "error" or "success"
}
