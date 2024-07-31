package ada

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tensved/bobrix/contracts"
	"log/slog"
	"net/url"
)

func NewADAService(adaHost string) *contracts.Service {

	return &contracts.Service{
		Name:        "ada",
		Description: "Ada is an AI language model trained by OpenAI.",
		Methods: map[string]*contracts.Method{
			"generate": {
				Name:        "generate",
				Description: "Generate text using Ada's language model.",
				Inputs: []*contracts.Input{
					{
						Name:        "prompt",
						Description: "The prompt to generate text from.",
						IsRequired:  true,
					},
				},
				Outputs: []*contracts.Output{
					{
						Name:        "text",
						Description: "The generated text.",
					},
				},
				Handler: NewADAHandler(adaHost),
			},
		},
	}
}

type ADAHandler struct {
	URL url.URL
}

func NewADAHandler(adaHost string) *ADAHandler {
	adaURL := url.URL{
		Scheme: "wss",
		Host:   adaHost,
		Path:   "/",
	}
	return &ADAHandler{
		URL: adaURL,
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

func (h *ADAHandler) Do(inputData map[string]any) *contracts.MethodResponse {
	conn, _, err := websocket.DefaultDialer.Dial(h.URL.String(), nil)

	if err != nil {
		slog.Error("failed to dial", "err", err)
		return &contracts.MethodResponse{
			Error: fmt.Errorf("failed to dial: %w", err),
		}
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("failed to close connection", "err", err)
		}
	}(conn)

	responseType := "text"

	if respType, ok := inputData["response_type"]; ok {
		responseType = respType.(string)
	}

	message := InputMessage{
		ClientName:   "bot",
		Username:     "username",
		RequestType:  "text",
		ResponseType: responseType,
	}

	audioData, ok := inputData["audio"]
	if ok {
		message.RequestType = "speech"

		audio, ok := audioData.(string)
		if !ok {
			return &contracts.MethodResponse{
				Error: fmt.Errorf("audio not found"),
			}
		}
		message.Speech = audio
	} else {

		textData, ok := inputData["prompt"]

		if !ok {
			return &contracts.MethodResponse{
				Error: fmt.Errorf("prompt not found"),
			}
		}

		message.RequestType = "text"
		message.Text = textData.(string)

	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		slog.Error("failed to marshal Input JSON", "err", err)
		return &contracts.MethodResponse{
			Error: fmt.Errorf("failed to marshal Input JSON: %w", err),
		}
	}

	slog.Error("sending message", "message", string(messageBytes))

	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		slog.Error("failed to send Input JSON", "err", err)
		return &contracts.MethodResponse{
			Error: fmt.Errorf("failed to send Input JSON: %w", err),
		}
	}

	_, p, err := conn.ReadMessage()
	if err != nil {
		slog.Error("failed to read OutputMessage JSON", "err", err)
		return &contracts.MethodResponse{
			Error: fmt.Errorf("failed to read OutputMessage JSON: %w", err),
		}
	}

	var responseMessage OutputMessage
	err = json.Unmarshal(p, &responseMessage)
	if err != nil {
		slog.Error("failed to unmarshal OutputMessage JSON", "err", err)
		return &contracts.MethodResponse{
			Error: fmt.Errorf("failed to unmarshal OutputMessage JSON: %w", err),
		}
	}

	slog.Info("Received new message from ADA", "answer", responseMessage.Answer, "RESPONSE", responseMessage)

	return &contracts.MethodResponse{
		Data: map[string]any{
			"answer": responseMessage.Answer,
		},
	}
}
