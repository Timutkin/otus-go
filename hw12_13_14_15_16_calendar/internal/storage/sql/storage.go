package sqlstorage

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

var (
	ErrEventIDAlreadyExist = errors.New("event id already exist")
	EmptyEvent             = storage.Event{}
)

type Storage struct {
	db        *sqlx.DB
	dsn       string
	tableName string
}

func (s *Storage) Create(ctx context.Context, e storage.Event) error {
	builder := sq.Insert(s.tableName).
		Columns("id", "title", "date_time", "event_duration", "description", "user_id", "notification_time").
		Values(e.ID, e.Title, e.DateTime, e.EventDuration, e.Description, e.UserID, e.NotificationTime)
	if e.NotificationTime != nil {
		builder = builder.Columns("notification_status").Values("PENDING")
	}
	sql, args, err := builder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("building create user query : %w", err)
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				return ErrEventIDAlreadyExist
			}
		}
		return fmt.Errorf("exec create user query : %w", err)
	}
	return err
}

func (s *Storage) Update(ctx context.Context, newEvent storage.Event) error {
	sql := sq.Update(s.tableName)
	if newEvent.UserID != nil {
		sql = sql.Set("user_id", newEvent.UserID)
	}
	if newEvent.Title != nil {
		sql = sql.Set("title", newEvent.Title)
	}
	if newEvent.DateTime != nil {
		sql = sql.Set("date_time", newEvent.DateTime)
	}
	if newEvent.EventDuration != nil {
		sql = sql.Set("event_duration", newEvent.EventDuration)
	}
	if newEvent.Description != nil {
		sql = sql.Set("description", newEvent.Description)
	}
	if newEvent.NotificationTime != nil {
		sql = sql.Set("notification_time", newEvent.NotificationTime)
	}
	if newEvent.NotificationStatus != nil {
		sql = sql.Set("notification_status", newEvent.NotificationStatus)
	}

	sql = sql.Where(sq.Eq{"id": newEvent.ID})
	query, args, err := sql.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("error while build update query %w", err)
	}
	_, err = s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Storage) Delete(ctx context.Context, eventID uuid.UUID) error {
	_, err := sq.Delete(s.tableName).
		Where(sq.Eq{"id": eventID}).
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db).
		ExecContext(ctx)
	return err
}

func (s *Storage) GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]storage.Event, error) {
	events := make([]storage.Event, 0)
	sql, args, err := sq.Select("*").From(s.tableName).Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return events, err
	}
	rows, err := s.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return events, fmt.Errorf("error while executing select * from events where user_id = $1 : %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return events, fmt.Errorf(ErrParsingToStructError, "storage.Event", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (s *Storage) GetByID(ctx context.Context, eventID uuid.UUID) (storage.Event, error) {
	sql, args, err := sq.Select("*").From(s.tableName).Where(sq.Eq{"id": eventID}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return EmptyEvent, err
	}
	rows, err := s.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return EmptyEvent, fmt.Errorf("error while executing select * from events where id = $1 : %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return EmptyEvent, fmt.Errorf(ErrParsingToStructError, "storage.Event", err)
		}
		return event, nil
	}
	return EmptyEvent, storage.ErrEventNotFoundErr
}

func (s *Storage) FindByCurrentTimeByMinutesAndPendingStatus() ([]storage.Event, error) {
	events := make([]storage.Event, 0)
	sql, args, err := sq.Select("*").From(s.tableName).Where(
		sq.And{
			sq.Eq{"notification_status": "PENDING"},
			sq.Expr("DATE_TRUNC('minute', notification_time) = DATE_TRUNC('minute', NOW())"),
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return events, err
	}
	rows, err := s.db.Queryx(sql, args...)
	if err != nil {
		return events, err
	}
	defer rows.Close()
	for rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return events, fmt.Errorf(ErrParsingToStructError, "storage.Event", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (s *Storage) FindByDateTimeMoreOrEqual(dateTime time.Time) ([]storage.Event, error) {
	events := make([]storage.Event, 0)
	sql, args, err := sq.Select("*").From(s.tableName).Where(sq.LtOrEq{"date_time": dateTime}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return events, err
	}
	rows, err := s.db.Queryx(sql, args...)
	if err != nil {
		return events, err
	}
	defer rows.Close()
	for rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return events, fmt.Errorf(ErrParsingToStructError, "storage.Event", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func New(cfg config.DBConf) *Storage {
	tables := cfg.Tables
	return &Storage{
		dsn:       cfg.CollectDsn(),
		tableName: tables.Schema + "." + "events",
	}
}

func (s *Storage) Connect(ctx context.Context) error {
	db, err := sqlx.ConnectContext(ctx, "pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to db : %w", err)
	}
	s.db = db
	return nil
}

func (s *Storage) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}
