package main

import (
	"encoding/json"
	mapset "github.com/deckarep/golang-set/v2"
	tdlib "github.com/zelenin/go-tdlib/client"
	"log"
	"os"
	"regexp"
)

type Config struct {
	ApiHash           string           `json:"api_hash"`
	ApiId             int32            `json:"api_id"`
	UseTestDc         bool             `json:"use_test_dc"`
	StateDir          string           `json:"state_dir"`
	LogVerbosityLevel int32            `json:"log_verbosity_level"`
	ForwardingConfig  ForwardingConfig `json:"forwarding_config"`
}

type ForwardingConfig struct {
	Source       []ParticipantConfig
	Destinations []ParticipantConfig
	Filter       RegexFilterConfig
	Forward      bool
}

type RegexFilterConfig struct {
	Regex string `json:"regex"`
}

type ForwardingConfigResolved struct {
	Source       mapset.Set[int64]
	Destinations []Participant
	Filter       MessageFilter
	Forward      bool
}

type MessageFilter interface {
	Passes(msg *tdlib.Message) bool
	Describe() string
}

type ParticipantConfig interface {
}

type ParticipantWithNameConfig struct {
	Username string `json:"username"`
}

type ParticipantWithIdConfig struct {
	ChatId int64 `json:"chat_id"`
}

type Participant struct {
	ChatId int64
	Name   string
}

func (p *ParticipantWithNameConfig) ParticipantType() string {
	return "name"
}

func (p *ParticipantWithIdConfig) ParticipantType() string {
	return "id"
}

func parseConfig() *Config {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "simple-telegram-forwarder.config.json"
	}
	bytes, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}

	if config.StateDir == "" {
		config.StateDir = "."
	}

	config.validate()
	return &config
}

func (config *Config) resolveForwardingConfig(client *tdlib.Client) *ForwardingConfigResolved {
	fc := config.ForwardingConfig
	var resolved ForwardingConfigResolved

	resolvedSources := make([]Participant, len(fc.Source))
	for i, source := range fc.Source {
		resolvedSources[i] = config.resolveParticipantConfig("source", client, source)
	}
	resolved.Source = mapset.NewSetWithSize[int64](len(resolvedSources))
	for _, source := range resolvedSources {
		resolved.Source.Add(source.ChatId)
	}

	resolved.Destinations = make([]Participant, len(fc.Destinations))
	for i, receiver := range fc.Destinations {
		resolved.Destinations[i] = config.resolveParticipantConfig("destination", client, receiver)
	}

	resolved.Forward = fc.Forward

	if fc.Filter.Regex != "" {
		r := regexp.MustCompile(fc.Filter.Regex)
		resolved.Filter = &RegexFilter{regex: r}
		log.Printf("Loaded filter %s", resolved.Filter.Describe())
	} else {
		resolved.Filter = &EmptyFilter{}
	}
	if fc.Forward {
		log.Printf("Will forward messages instead of sending a copy")
	}
	return &resolved
}

func (config *Config) resolveParticipantConfig(participantType string, client *tdlib.Client, pc ParticipantConfig) Participant {
	idConfig, ok := pc.(*ParticipantWithIdConfig)
	if !ok {
		name := pc.(*ParticipantWithNameConfig).Username
		chat, err := client.SearchPublicChat(&tdlib.SearchPublicChatRequest{Username: name})
		if err == nil {
			log.Printf(
				"Resolved %s participant with name='%s' to a chat with title='%s', chatId=%d\n",
				participantType, name, chat.Title, chat.Id)
			return Participant{ChatId: chat.Id, Name: chat.Title}
		}
		log.Fatalf("Could not find chat for username '%s'. %v", name, err)
	}
	chat, err := client.GetChat(&tdlib.GetChatRequest{ChatId: idConfig.ChatId})
	if err != nil {
		log.Fatalf("Could not find chat with id=%d. %v", idConfig.ChatId, err)
	} else {
		log.Printf("Resolved %s participant with chatId=%d to a chat with title='%s'\n",
			participantType, idConfig.ChatId, chat.Title)
	}
	return Participant{ChatId: idConfig.ChatId, Name: chat.Title}
}

func (config *Config) validate() {
	if config.ForwardingConfig.Source == nil || len(config.ForwardingConfig.Source) == 0 {
		log.Fatalln("Missing forwarding_config.source")
	}
	if len(config.ForwardingConfig.Destinations) == 0 {
		log.Fatalln("forwarding_config.destinations is empty")
	}
}
