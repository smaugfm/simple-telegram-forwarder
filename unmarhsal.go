package main

import (
	"encoding/json"
)

func (forwardConfig *ForwardingConfig) UnmarshalJSON(data []byte) error {
	var tmp struct {
		Source       json.RawMessage   `json:"source"`
		Destinations []json.RawMessage `json:"destinations"`
		Filter       RegexFilterConfig `json:"filter"`
		Forward      bool              `json:"forward"`
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	sender, err := unmarshalParticipantConfig(tmp.Source)
	if err != nil {
		return err
	}
	forwardConfig.Source = *sender

	forwardConfig.Destinations = make([]ParticipantConfig, len(tmp.Destinations))
	for i, receiverRaw := range tmp.Destinations {
		receiver, err := unmarshalParticipantConfig(receiverRaw)
		if err != nil {
			return err
		}
		forwardConfig.Destinations[i] = *receiver
	}
	forwardConfig.Filter = tmp.Filter
	forwardConfig.Forward = tmp.Forward
	return nil
}

func unmarshalParticipantConfig(data json.RawMessage) (*ParticipantConfig, error) {
	var participant ParticipantConfig
	var mapJson map[string]any
	err := json.Unmarshal(data, &mapJson)
	if err != nil {
		return nil, err
	}

	if mapJson["username"] != nil {
		var senderName ParticipantWithNameConfig
		err = json.Unmarshal(data, &senderName)
		if err != nil {
			return nil, err
		}
		participant = &senderName
	} else if mapJson["chat_id"] != nil {
		var senderId ParticipantWithIdConfig
		err = json.Unmarshal(data, &senderId)
		if err != nil {
			return nil, err
		}
		participant = &senderId
	}
	return &participant, nil
}
