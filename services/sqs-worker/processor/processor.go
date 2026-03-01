package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fuvekonse/sqs-worker/jobmsg"
	"fuvekonse/sqs-worker/repo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProcessTicketJob processes one ticket job message and writes to the database.
func ProcessTicketJob(ctx context.Context, db *gorm.DB, body []byte) error {
	var msg jobmsg.TicketJobMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}

	tr := repo.NewTicketRepo(db)

	switch msg.Action {
	case jobmsg.ActionPurchaseTicket:
		uid, err := uuid.Parse(msg.UserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		tid, err := uuid.Parse(msg.TierID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		_, err = tr.PurchaseTicket(ctx, uid, tid)
		return err
	case jobmsg.ActionConfirmPayment:
		uid, err := uuid.Parse(msg.UserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		existing, err := tr.GetUserTicket(ctx, uid)
		if err != nil {
			return err
		}
		if existing == nil {
			return ErrNoTicketFound
		}
		_, err = tr.ConfirmPayment(ctx, existing.Id, uid)
		return err
	case jobmsg.ActionCancelTicket:
		uid, err := uuid.Parse(msg.UserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		existing, err := tr.GetUserTicket(ctx, uid)
		if err != nil {
			return err
		}
		if existing == nil {
			return ErrNoTicketFound
		}
		return tr.CancelTicket(ctx, existing.Id, uid)
	case jobmsg.ActionUpdateBadge:
		uid, err := uuid.Parse(msg.UserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		existing, err := tr.GetUserTicket(ctx, uid)
		if err != nil {
			return err
		}
		if existing == nil {
			return ErrNoTicketFound
		}
		_, err = tr.UpdateBadgeDetails(ctx, existing.Id, uid, msg.ConBadgeName, msg.BadgeImage, msg.IsFursuiter, msg.IsFursuitStaff)
		return err
	case jobmsg.ActionApproveTicket:
		tid, err := uuid.Parse(msg.TicketID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		sid, err := uuid.Parse(msg.StaffID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		_, err = tr.ApproveTicket(ctx, tid, sid)
		return err
	case jobmsg.ActionDenyTicket:
		tid, err := uuid.Parse(msg.TicketID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		sid, err := uuid.Parse(msg.StaffID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		_, err = tr.DenyTicket(ctx, tid, sid, msg.Reason)
		return err
	case jobmsg.ActionUpgradeTicket:
		uid, err := uuid.Parse(msg.UserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		tid, err := uuid.Parse(msg.TierID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		_, err = tr.UpgradeTicketTier(ctx, uid, tid)
		return err
	case jobmsg.ActionBlacklistUser:
		uid, err := uuid.Parse(msg.TargetUserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		return tr.BlacklistUser(ctx, uid, msg.Reason)
	case jobmsg.ActionUnblacklistUser:
		uid, err := uuid.Parse(msg.TargetUserID)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidUUID, err)
		}
		return tr.UnblacklistUser(ctx, uid)
	default:
		return ErrUnknownAction
	}
}

var (
	ErrNoTicketFound = errors.New("no ticket found for this user")
	ErrUnknownAction = errors.New("unknown ticket job action")
	ErrInvalidUUID   = errors.New("invalid UUID format")
)

// IsPermanentError returns true if retrying the message won't fix the error.
func IsPermanentError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, repo.ErrTicketTierNotFound) ||
		errors.Is(err, repo.ErrTicketNotFound) ||
		errors.Is(err, repo.ErrOutOfStock) ||
		errors.Is(err, repo.ErrUserAlreadyHasTicket) ||
		errors.Is(err, repo.ErrUserBlacklisted) ||
		errors.Is(err, repo.ErrInvalidTicketStatus) ||
		errors.Is(err, repo.ErrCannotDowngrade) ||
		errors.Is(err, repo.ErrTicketDenied) ||
		errors.Is(err, ErrInvalidUUID) ||
		errors.Is(err, ErrNoTicketFound) ||
		errors.Is(err, ErrUnknownAction) ||
		errors.Is(err, gorm.ErrRecordNotFound)
}
