package main

import (
	"encoding/json"
)

type internalConfig struct {
	Sources      []json.RawMessage `json:"sources"`
	Destinations []json.RawMessage `json:"destinations"`
	Filter       RegexFilterConfig `json:"filter"`
	Forward      bool              `json:"forward"`
}

func (forwardConfig *ForwardingConfig) UnmarshalJSON(data []byte) error {
	var tmp internalConfig

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	forwardConfig.Sources, err = unmarshalParticipantConfigArray(tmp.Sources)
	if err != nil {
		return err
	}

	forwardConfig.Destinations, err = unmarshalParticipantConfigArray(tmp.Destinations)
	if err != nil {
		return err
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

func unmarshalParticipantConfigArray(array []json.RawMessage) ([]ParticipantConfig, error) {
	result := make([]ParticipantConfig, len(array))
	for i, receiverRaw := range array {
		receiver, err := unmarshalParticipantConfig(receiverRaw)
		if err != nil {
			return nil, err
		}
		result[i] = *receiver
	}
	return result, nil
}
