package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type callbackAction string

const (
	actionApprove callbackAction = "approve"
	actionAdmin   callbackAction = "admin"
)

func (b *Bot) handleCallback(ctx context.Context, query *tgbotapi.CallbackQuery) error {
	callback := tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID)
	_, err := b.api.Request(callback)
	if err != nil {
		return fmt.Errorf("error sending callback request: %w", err)
	}

	var action map[callbackAction]any
	if err := json.Unmarshal([]byte(query.Data), &action); err != nil {
		return fmt.Errorf("error unmarshal action: %w", err)
	}

	if action[actionApprove] != nil {
		err := b.handleApproveCallback(ctx, action[actionApprove], query.Message.Chat.ID)
		if err != nil {
			return fmt.Errorf("error handle approve callback action: %w", err)
		}
	}
	if action[actionAdmin] != nil {
		err := b.handleAdminCallback(ctx, action[actionAdmin], query.Message.Chat.ID)
		if err != nil {
			return fmt.Errorf("error handle admin callback action: %w", err)
		}
	}

	return nil
}

func (b *Bot) handleApproveCallback(ctx context.Context, actionData any, chatID int64) error {
	userChatID, err := getChatIDFromData(actionData)
	if err != nil {
		return err
	}
	user, err := b.repo.User(ctx, userChatID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	user.Approved = !user.Approved
	if err = b.repo.UserSave(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	var answer string
	if user.Approved {
		answer = fmt.Sprintf("%s User %s approved!", emojiApproved, user)
		b.SendMessage(ctx, user.ChatID, emojiApproved+" You are approved!", nil)
	} else {
		answer = fmt.Sprintf("%s User %s denied!", emojiDeclined, user)
	}

	// send message to admin
	b.SendMessage(ctx, chatID, answer, nil)
	return nil
}

func (b *Bot) handleAdminCallback(ctx context.Context, actionData any, chatID int64) error {
	userChatID, err := getChatIDFromData(actionData)
	if err != nil {
		return err
	}
	user, err := b.repo.User(ctx, userChatID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	user.Admin = !user.Admin

	// check if user was a last admin
	admins, err := b.repo.Admins(ctx)
	if err != nil {
		return fmt.Errorf("error getting admins: %w", err)
	}
	if !user.Admin && len(admins) == 1 && admins[0].ChatID == user.ChatID {
		b.SendMessage(ctx, chatID,
			emojiAlert+" You are the last admin, you can't remove admin rights from yourself", nil)
		return nil
	}

	if err := b.repo.UserSave(ctx, user); err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}
	var answer string
	if user.Admin {
		answer = fmt.Sprintf("%s User %s is admin now!", emojiAdmin, user)
		b.SendMessage(ctx, user.ChatID, "👑 You are now an admin!", nil)
	} else {
		answer = fmt.Sprintf("%s User %s is not admin anymore!", emojiUser, user)
	}
	b.SendMessage(ctx, chatID, answer, nil)
	return nil
}

func getChatIDFromData(actionData any) (int64, error) {
	userChatIDFloat64, ok := actionData.(float64) // json marshaled as float64!
	if !ok {
		return 0, errors.New("error to parse chat id")
	}
	return int64(userChatIDFloat64), nil
}
