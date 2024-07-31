package bobrix

import (
	"maunium.net/go/mautrix/event"
	"regexp"
)

func DefaultContractParser() func(evt *event.Event) *AIRequest {
	reMsg := regexp.MustCompile(
		`(@(?P<bot>\w+)\s+)*(-service:(?P<service>\w+)\s+)*(-method:(?P<method>\w+)\s)*(?P<inputs>.*)`,
	)

	reInputs := regexp.MustCompile(`-(\w+):"((?:\\"|[^"])*)"`)

	return func(evt *event.Event) *AIRequest {

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

		return &AIRequest{
			ServiceName: groups["service"],
			MethodName:  groups["method"],
			InputParams: inputsData,
		}
	}

}
