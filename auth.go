package main

import (
	"flag"
	tdlib "github.com/zelenin/go-tdlib/client"
	"log"
	"os"
	"path/filepath"
)

func (config *Config) authorize() *tdlib.Client {
	var authOnly = flag.Bool("auth-only", false, "Only authorize to Telegram and then exit")
	flag.Parse()

	authorizer := tdlib.ClientAuthorizer()
	go tdlib.CliInteractor(authorizer)
	authorizer.TdlibParameters <- &tdlib.SetTdlibParametersRequest{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(config.StateDir, ".tdlib", "database"),
		FilesDirectory:         filepath.Join(config.StateDir, ".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  config.ApiId,
		ApiHash:                config.ApiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	_, err := tdlib.SetLogVerbosityLevel(&tdlib.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: config.LogVerbosityLevel,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %v", err)
	}

	client, err := tdlib.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %v", err)
	}

	optionValue, err := tdlib.GetOption(&tdlib.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %v", err)
	}

	log.Printf("TDLib version: %s", optionValue.(*tdlib.OptionValueString).Value)

	me, err := client.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %v", err)
	}

	log.Printf("Authorized user: %s %s [%s]", me.FirstName, me.LastName, me.Usernames.ActiveUsernames[0])

	if *authOnly {
		os.Exit(0)
		return nil
	}

	return client
}
