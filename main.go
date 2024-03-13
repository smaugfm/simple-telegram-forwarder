package main

import (
	tdlib "github.com/zelenin/go-tdlib/client"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config := parseConfig()
	client := config.authorize()
	resolved := config.resolveForwardingConfig(client)

	listener := client.GetListener()
	defer listener.Close()

	exitOnSigTerm(client)

	log.Println("Listening for incoming messages...")
	for update := range listener.Updates {
		if update.GetType() == tdlib.TypeUpdateNewMessage {
			msg := update.(*tdlib.UpdateNewMessage).Message
			processMessage(client, resolved, msg)
		}
	}
}

func exitOnSigTerm(client *tdlib.Client) {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.Stop()
		os.Exit(1)
	}()
}
