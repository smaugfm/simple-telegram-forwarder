package main

import (
	"fmt"
	"github.com/samber/lo"
	tdlib "github.com/zelenin/go-tdlib/client"
	"go/types"
	"log"
)

func processMessage(client *tdlib.Client, config *ForwardingConfigResolved, msg *tdlib.Message) {
	if !config.Sources.Contains(msg.ChatId) {
		return
	}
	logIncomingMessage(client, msg)
	inputContent, err := makeInputMessageContent(msg.Content)
	if err != nil {
		log.Print(err)
		return
	}
	if !config.Filter.Passes(msg) {
		log.Printf("Did not pass filter %s", config.Filter.Describe())
		return
	}
	for _, destination := range config.Destinations {
		if config.Forward {
			log.Printf("Forwarding to '%s' (%d)", destination.Name, destination.ChatId)
			_, err = client.ForwardMessages(&tdlib.ForwardMessagesRequest{
				ChatId:     destination.ChatId,
				FromChatId: msg.ChatId,
				MessageIds: []int64{msg.Id},
			})
		} else {
			log.Printf("Sending to '%s' (%d)", destination.Name, destination.ChatId)
			_, err = client.SendMessage(&tdlib.SendMessageRequest{
				ChatId:              destination.ChatId,
				InputMessageContent: inputContent,
			})
		}
		if err != nil {
			log.Printf("Failed to send to '%s' (%d). %v", destination.Name, destination.ChatId, err)
			continue
		}
	}
}

func makeInputMessageContent(content tdlib.MessageContent) (tdlib.InputMessageContent, error) {
	switch content.MessageContentType() {
	case tdlib.TypeMessageText:
		c := content.(*tdlib.MessageText)
		return &tdlib.InputMessageText{Text: c.Text, LinkPreviewOptions: c.LinkPreviewOptions}, nil
	case tdlib.TypeMessageAnimation:
		c := content.(*tdlib.MessageAnimation)
		file, err := getInputFile(c.Animation.Animation)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.Animation.Thumbnail)
		return &tdlib.InputMessageAnimation{
			Animation:  file,
			Thumbnail:  thumbnail,
			Duration:   c.Animation.Duration,
			Width:      c.Animation.Width,
			Height:     c.Animation.Height,
			Caption:    c.Caption,
			HasSpoiler: c.HasSpoiler,
		}, nil
	case tdlib.TypeMessageAudio:
		c := content.(*tdlib.MessageAudio)
		file, err := getInputFile(c.Audio.Audio)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.Audio.AlbumCoverThumbnail)
		return &tdlib.InputMessageAudio{
			Audio:               file,
			AlbumCoverThumbnail: thumbnail,
			Duration:            c.Audio.Duration,
			Title:               c.Audio.Title,
			Performer:           c.Audio.Performer,
			Caption:             c.Caption,
		}, nil
	case tdlib.TypeMessageDocument:
		c := content.(*tdlib.MessageDocument)
		file, err := getInputFile(c.Document.Document)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.Document.Thumbnail)
		return &tdlib.InputMessageDocument{
			Document:  file,
			Thumbnail: thumbnail,
			Caption:   c.Caption,
		}, nil
	case tdlib.TypeMessagePhoto:
		c := content.(*tdlib.MessagePhoto)
		maxSize := lo.MaxBy(c.Photo.Sizes, func(i *tdlib.PhotoSize, max *tdlib.PhotoSize) bool {
			return i.Photo.Size > max.Photo.Size
		})
		minSize := lo.MinBy(c.Photo.Sizes, func(i *tdlib.PhotoSize, max *tdlib.PhotoSize) bool {
			return i.Photo.Size < max.Photo.Size
		})
		file, err := getInputFile(maxSize.Photo)
		if err != nil {
			return nil, err
		}
		size, err := getInputThumbnailFromPhotoSize(minSize)
		return &tdlib.InputMessagePhoto{
			Photo:      file,
			Thumbnail:  size,
			Width:      maxSize.Width,
			Height:     maxSize.Height,
			Caption:    c.Caption,
			HasSpoiler: c.HasSpoiler,
		}, nil
	case tdlib.TypeMessageExpiredPhoto:
		return nil, types.Error{Msg: "Photo has expired"}
	case tdlib.TypeMessageSticker:
		c := content.(*tdlib.MessageSticker)
		file, err := getInputFile(c.Sticker.Sticker)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.Sticker.Thumbnail)
		return &tdlib.InputMessageSticker{
			Sticker:   file,
			Thumbnail: thumbnail,
			Width:     c.Sticker.Width,
			Height:    c.Sticker.Height,
			Emoji:     c.Sticker.Emoji,
		}, nil
	case tdlib.TypeMessageVideo:
		c := content.(*tdlib.MessageVideo)
		file, err := getInputFile(c.Video.Video)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.Video.Thumbnail)
		return &tdlib.InputMessageVideo{
			Video:             file,
			Thumbnail:         thumbnail,
			Duration:          c.Video.Duration,
			Height:            c.Video.Height,
			Width:             c.Video.Width,
			SupportsStreaming: c.Video.SupportsStreaming,
			Caption:           c.Caption,
			HasSpoiler:        c.HasSpoiler,
		}, nil
	case tdlib.TypeMessageExpiredVideo:
		return nil, types.Error{Msg: "Video has expired"}
	case tdlib.TypeMessageVideoNote:
		c := content.(*tdlib.MessageVideoNote)
		file, err := getInputFile(c.VideoNote.Video)
		if err != nil {
			return nil, err
		}
		thumbnail, err := getInputThumbnail(c.VideoNote.Thumbnail)
		return &tdlib.InputMessageVideoNote{
			VideoNote: file,
			Thumbnail: thumbnail,
			Duration:  c.VideoNote.Duration,
			Length:    c.VideoNote.Length,
		}, nil
	case tdlib.TypeMessageVoiceNote:
		c := content.(*tdlib.MessageVoiceNote)
		file, err := getInputFile(c.VoiceNote.Voice)
		if err != nil {
			return nil, err
		}
		return &tdlib.InputMessageVoiceNote{
			VoiceNote: file,
			Duration:  c.VoiceNote.Duration,
			Waveform:  c.VoiceNote.Waveform,
			Caption:   c.Caption,
		}, nil
	case tdlib.TypeMessageLocation:
		c := content.(*tdlib.MessageLocation)
		return &tdlib.InputMessageLocation{
			Location:             c.Location,
			LivePeriod:           c.LivePeriod,
			Heading:              c.Heading,
			ProximityAlertRadius: c.ProximityAlertRadius,
		}, nil
	case tdlib.TypeMessageVenue:
		c := content.(*tdlib.MessageVenue)
		return &tdlib.InputMessageVenue{
			Venue: c.Venue,
		}, nil
	case tdlib.TypeMessageContact:
		c := content.(*tdlib.MessageContact)
		return &tdlib.InputMessageContact{
			Contact: c.Contact,
		}, nil
	case tdlib.TypeMessageDice:
		c := content.(*tdlib.MessageDice)
		return &tdlib.InputMessageDice{
			Emoji: c.Emoji,
		}, nil
	case tdlib.TypeMessagePoll:
		c := content.(*tdlib.MessagePoll)
		return &tdlib.InputMessagePoll{
			Question: c.Poll.Question,
			Options: lo.Map(c.Poll.Options, func(item *tdlib.PollOption, index int) string {
				return item.Text
			}),
			IsAnonymous: c.Poll.IsAnonymous,
			Type:        c.Poll.Type,
			OpenPeriod:  c.Poll.OpenPeriod,
			CloseDate:   c.Poll.CloseDate,
			IsClosed:    c.Poll.IsClosed,
		}, nil
	case tdlib.TypeMessageStory:
		c := content.(*tdlib.MessageStory)
		return &tdlib.InputMessageStory{
			StorySenderChatId: c.StorySenderChatId,
			StoryId:           c.StoryId,
		}, nil
	case tdlib.TypeMessageAnimatedEmoji:
	case tdlib.TypeMessageGame:
	case tdlib.TypeMessageInvoice:
	case tdlib.TypeMessageCall:
	case tdlib.TypeMessageVideoChatScheduled:
	case tdlib.TypeMessageVideoChatStarted:
	case tdlib.TypeMessageVideoChatEnded:
	case tdlib.TypeMessageInviteVideoChatParticipants:
	case tdlib.TypeMessageBasicGroupChatCreate:
	case tdlib.TypeMessageSupergroupChatCreate:
	case tdlib.TypeMessageChatChangeTitle:
	case tdlib.TypeMessageChatChangePhoto:
	case tdlib.TypeMessageChatDeletePhoto:
	case tdlib.TypeMessageChatAddMembers:
	case tdlib.TypeMessageChatJoinByLink:
	case tdlib.TypeMessageChatJoinByRequest:
	case tdlib.TypeMessageChatDeleteMember:
	case tdlib.TypeMessageChatUpgradeTo:
	case tdlib.TypeMessageChatUpgradeFrom:
	case tdlib.TypeMessagePinMessage:
	case tdlib.TypeMessageScreenshotTaken:
	case tdlib.TypeMessageChatSetBackground:
	case tdlib.TypeMessageChatSetTheme:
	case tdlib.TypeMessageChatSetMessageAutoDeleteTime:
	case tdlib.TypeMessageForumTopicCreated:
	case tdlib.TypeMessageForumTopicEdited:
	case tdlib.TypeMessageForumTopicIsClosedToggled:
	case tdlib.TypeMessageForumTopicIsHiddenToggled:
	case tdlib.TypeMessageSuggestProfilePhoto:
	case tdlib.TypeMessageCustomServiceAction:
	case tdlib.TypeMessageGameScore:
	case tdlib.TypeMessagePaymentSuccessful:
	case tdlib.TypeMessagePaymentSuccessfulBot:
	case tdlib.TypeMessageGiftedPremium:
	case tdlib.TypeMessagePremiumGiftCode:
	case tdlib.TypeMessagePremiumGiveawayCreated:
	case tdlib.TypeMessagePremiumGiveaway:
	case tdlib.TypeMessagePremiumGiveawayCompleted:
	case tdlib.TypeMessageContactRegistered:
	case tdlib.TypeMessageUserShared:
	case tdlib.TypeMessageChatShared:
	case tdlib.TypeMessageBotWriteAccessAllowed:
	case tdlib.TypeMessageWebAppDataSent:
	case tdlib.TypeMessageWebAppDataReceived:
	case tdlib.TypeMessagePassportDataSent:
	case tdlib.TypeMessagePassportDataReceived:
	case tdlib.TypeMessageProximityAlertTriggered:
	case tdlib.TypeMessageUnsupported:
		return nil, notSupportedError(content.MessageContentType())
	}
	return nil, types.Error{Msg: fmt.Sprintf("Unknown message type %s", content.MessageContentType())}
}

func getTextFromMessage(msg *tdlib.Message) string {

	switch msg.Content.MessageContentType() {
	case tdlib.TypeMessageText:
		return msg.Content.(*tdlib.MessageText).Text.Text
	case tdlib.TypeMessageAnimation:
		return msg.Content.(*tdlib.MessageAnimation).Caption.Text
	case tdlib.TypeMessageAudio:
		return msg.Content.(*tdlib.MessageAudio).Caption.Text
	case tdlib.TypeMessageDocument:
		return msg.Content.(*tdlib.MessageDocument).Caption.Text
	case tdlib.TypeMessagePhoto:
		return msg.Content.(*tdlib.MessagePhoto).Caption.Text
	case tdlib.TypeMessageVideo:
		return msg.Content.(*tdlib.MessageVideo).Caption.Text
	case tdlib.TypeMessageVenue:
		return msg.Content.(*tdlib.MessageVenue).Venue.Title
	case tdlib.TypeMessageContact:
		return fmt.Sprintf("%s %s", msg.Content.(*tdlib.MessageContact).Contact.FirstName, msg.Content.(*tdlib.MessageContact).Contact.LastName)
	}
	return ""
}

func notSupportedError(msgType string) types.Error {
	return types.Error{Msg: fmt.Sprintf("Message type %s is not supported. Skipping...", msgType)}
}

func getInputFile(file *tdlib.File) (*tdlib.InputFileId, error) {
	if file == nil {
		return nil, types.Error{Msg: "File is nil"}
	}
	return &tdlib.InputFileId{Id: file.Id}, nil
}

func getInputThumbnail(thumbnail *tdlib.Thumbnail) (*tdlib.InputThumbnail, error) {
	if thumbnail == nil {
		return nil, types.Error{Msg: "Thumbnail is nil"}
	}
	return &tdlib.InputThumbnail{
		Thumbnail: &tdlib.InputFileId{Id: thumbnail.File.Id},
		Height:    thumbnail.Height,
		Width:     thumbnail.Width,
	}, nil
}

func getInputThumbnailFromPhotoSize(photoSize *tdlib.PhotoSize) (*tdlib.InputThumbnail, error) {
	if photoSize == nil || photoSize.Photo == nil {
		return nil, types.Error{Msg: "PhotoSize in nil"}
	}
	return &tdlib.InputThumbnail{
		Thumbnail: &tdlib.InputFileId{Id: photoSize.Photo.Id},
		Height:    photoSize.Height,
		Width:     photoSize.Width,
	}, nil
}

func logIncomingMessage(client *tdlib.Client, msg *tdlib.Message) bool {
	chat, err := getChatByChatId(client, msg.ChatId)
	if err != nil {
		log.Printf("New message %s %d in chat %d but failed to get chat info",
			msg.Content.MessageContentType(), msg.Id, msg.ChatId)
		return true
	}
	text := getTextFromMessage(msg)
	var sender string
	var senderId int64
	if msg.SenderId.MessageSenderType() == tdlib.TypeMessageSenderUser {
		senderId = msg.SenderId.(*tdlib.MessageSenderUser).UserId
		user, err := client.GetUser(&tdlib.GetUserRequest{UserId: senderId})
		if err != nil {
			log.Printf("New message %s %d in chat %d from %d but failed to get sender user info\n%s",
				msg.Content.MessageContentType(),
				msg.Id,
				msg.ChatId,
				senderId,
				text,
			)
			return true
		}
		sender = fmt.Sprintf("%s %s [%s]", user.FirstName, user.LastName, user.Usernames.ActiveUsernames[0])
	} else {
		senderId = msg.SenderId.(*tdlib.MessageSenderChat).ChatId
		senderChat, err := getChatByChatId(client, senderId)
		if err != nil {
			log.Printf("New message %s %d in chat %d from %d but failed to get sender chat info\n%s",
				msg.Content.MessageContentType(),
				msg.Id,
				msg.ChatId,
				senderId,
				text,
			)
			return true
		}
		sender = senderChat.Title
	}
	log.Printf("New message %s %d in %s (%d) from %s (%d)\n%s",
		msg.Content.MessageContentType(),
		msg.Id,
		chat.Title,
		msg.ChatId,
		sender,
		senderId,
		text)
	return false
}

func getChatByChatId(client *tdlib.Client, chatId int64) (*tdlib.Chat, error) {
	return client.GetChat(&tdlib.GetChatRequest{ChatId: chatId})
}
