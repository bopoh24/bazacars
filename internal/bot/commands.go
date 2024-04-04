package bot

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const greetingMessage = `
Greetings, car enthusiast! I am your trusted assistant in finding the best deals on cars for sale. 
With my database, you'll always be in the loop on the most exciting listings. 

Trust me and find your dream car hassle-free!`

func (b *Bot) commandHelpHandler(ctx context.Context, chatID int64) {
	b.SendMessage(ctx, chatID, greetingMessage, nil)
}

func (b *Bot) commandUsersHandler(ctx context.Context, chatID int64) {
	users, err := b.repo.Users(ctx)
	if err != nil {
		b.logger.Error("Error getting users", "err", err)
		return
	}
	userList := "<strong>Users:</strong>\n"
	for _, user := range users {
		// emoji done there
		if user.Approved {
			userList += emojiApproved + " "
		} else {
			userList += emojiDeclined + " "
		}
		userList += user.String()
		userList += "\n"
	}
	b.SendMessage(ctx, chatID, userList, nil)
}

func (b *Bot) commandApproveHandler(ctx context.Context, chatID int64) {
	// admin only
	if !b.isUserAdmin(ctx, chatID) {
		b.SendMessage(ctx, chatID, "You are not admin", nil)
		return
	}
	// all users not admins
	users, err := b.repo.Users(ctx)
	if err != nil {
		b.logger.Error("Error getting users", "err", err)
		return
	}

	buttonRows := make([][]tgbotapi.InlineKeyboardButton, 0)
	for _, user := range users {
		if user.Admin {
			continue
		}
		emoji := emojiDeclined
		if user.Approved {
			emoji = emojiApproved
		}
		btnText := fmt.Sprintf("%s %s", emoji, user)

		action := map[callbackAction]int64{
			actionApprove: user.ChatID,
		}
		res, err := json.Marshal(action)
		if err != nil {
			b.logger.Error("Error marshal action", "err", err)
			continue
		}

		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, string(res)),
		))
	}
	if len(buttonRows) == 0 {
		b.SendMessage(ctx, chatID, "No users to approve/deny", nil)
		return
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
	b.SendMessage(ctx, chatID, "Select user to approve/deny", &keyboard)
}

func (b *Bot) commandAdminsHandler(ctx context.Context, chatID int64) {
	// admin only
	if !b.isUserAdmin(ctx, chatID) {
		b.SendMessage(ctx, chatID, "You are not admin", nil)
		return
	}
	// all approved users
	users, err := b.repo.Users(ctx)
	if err != nil {
		b.logger.Error("Error getting users", "err", err)
		return
	}

	buttonRows := make([][]tgbotapi.InlineKeyboardButton, 0)
	for _, user := range users {
		if !user.Approved {
			continue
		}
		emoji := emojiUser
		if user.Admin {
			emoji = emojiAdmin
		}
		btnText := fmt.Sprintf("%s %s", emoji, user)

		action := map[callbackAction]int64{
			actionAdmin: user.ChatID,
		}
		res, err := json.Marshal(action)
		if err != nil {
			b.logger.Error("Error marshal action", "err", err)
			continue
		}

		buttonRows = append(buttonRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnText, string(res)),
		))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttonRows...)
	b.SendMessage(ctx, chatID, "Select user to make admin or remove admin", &keyboard)
}

func (b *Bot) isUserAdmin(ctx context.Context, chatID int64) bool {
	user, err := b.repo.User(ctx, chatID)
	if err != nil {
		b.logger.Error("Error getting user", "err", err)
		return false
	}
	return user.Admin
}
