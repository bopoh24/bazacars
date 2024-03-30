package bot

import (
	"context"
	"errors"
	"fmt"
	"github.com/bopoh24/bazacars/internal/model"
	"github.com/bopoh24/bazacars/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"time"
)

const (
	emojiApproved = "✅"
	emojiDeclined = "❌"
	emojiAdmin    = "👑"
	emojiUser     = "👤"
	emojiAlert    = "🚨"

	commandStart   = "start"
	commandHelp    = "help"
	commandUsers   = "users"
	commandApprove = "approve"
	commandAdmins  = "admins"
)

// Bot is a telegram bot
type Bot struct {
	api    *tgbotapi.BotAPI
	repo   repository.Repository
	logger *slog.Logger
}

// New returns new bot
func New(token string, repo repository.Repository, logger *slog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("telegram bot: %w", err)
	}
	logger = logger.With(slog.String("bot", api.Self.UserName))

	return &Bot{
		api:    api,
		logger: logger,
		repo:   repo,
	}, nil
}

// Run starts the bot
func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.CallbackQuery != nil {
				err := b.handleCallback(ctx, update.CallbackQuery)
				if err != nil {
					b.logger.Error("Error handling callback", "err", err,
						"chat_id", update.CallbackQuery.Message.Chat.ID)
					b.SendMessage(ctx, update.CallbackQuery.Message.Chat.ID,
						"Error handling callback. Try again...", nil)
				}
				continue
			}

			if update.Message == nil {
				continue
			}

			username := b.userInfoFromChat(update.Message.Chat)
			approved, err := b.isUserApproved(ctx, update.Message.Chat)
			if err != nil {
				b.logger.Error("Error checking user approval", "err", err,
					"chat_id", update.Message.Chat.ID, "username", username)
				continue
			}
			if !approved {
				b.logger.Warn("Not in a white list!", "chat_id", update.Message.Chat.ID, "username", username)
				b.SendMessage(ctx, update.Message.Chat.ID,
					"You are added to the waiting list. Wait for the administrator to approve you.", nil)
				continue
			}

			// Handle commands
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case commandStart, commandHelp:
					b.commandHelpHandler(ctx, update.Message.Chat.ID)
				case commandUsers:
					b.commandUsersHandler(ctx, update.Message.Chat.ID)
				case commandApprove:
					b.commandApproveHandler(ctx, update.Message.Chat.ID)
				case commandAdmins:
					b.commandAdminsHandler(ctx, update.Message.Chat.ID)
				default:
					b.SendMessage(ctx, update.Message.Chat.ID, "I don't know that command", nil)
				}
			}
		}
	}
}

// SendMessage sends message to chat
func (b *Bot) SendMessage(_ context.Context, chatID int64, text string, keyboard *tgbotapi.InlineKeyboardMarkup) {
	msgConf := tgbotapi.NewMessage(chatID, text)
	msgConf.ParseMode = "markdown"
	if keyboard != nil {
		msgConf.ReplyMarkup = keyboard
	}
	_, err := b.api.Send(msgConf)
	if err != nil {
		b.logger.Error("Error sending message", "err", err)
	}
}

func (b *Bot) userInfoFromChat(chat *tgbotapi.Chat) string {
	if chat.UserName != "" {
		return chat.UserName
	}
	if chat.FirstName != "" && chat.LastName != "" {
		return chat.FirstName + " " + chat.LastName
	}
	return fmt.Sprintf("id:%d", chat.ID)
}

func (b *Bot) isUserApproved(ctx context.Context, chat *tgbotapi.Chat) (bool, error) {
	user, err := b.repo.User(ctx, chat.ID)
	if errors.Is(err, repository.ErrNotFound) {
		// send message to admins
		admins, err := b.repo.Admins(ctx)
		if err != nil {
			return false, fmt.Errorf("error getting admins: %w", err)
		}

		// id admin list is empty, add user to db
		if len(admins) == 0 {
			return true, b.addUser(ctx, chat, true, true)
		}

		for _, admin := range admins {
			b.SendMessage(ctx, admin.ChatID,
				fmt.Sprintf("⚠️ Not in a white list!\n%s, chatID: %d", b.userInfoFromChat(chat), chat.ID), nil)
		}
		return false, b.addUser(ctx, chat, false, false)
	}
	return user.Approved, err
}

func (b *Bot) addUser(ctx context.Context, chat *tgbotapi.Chat, isAdmin, approved bool) error {
	err := b.repo.UserAdd(ctx, model.User{
		ChatID:    chat.ID,
		FirstName: chat.FirstName,
		LastName:  chat.LastName,
		Username:  chat.UserName,
		Admin:     isAdmin,
		Approved:  approved,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("error adding user: %w", err)
	}
	return nil
}
