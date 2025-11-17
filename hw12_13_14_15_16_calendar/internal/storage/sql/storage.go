package sqlstorage

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/config"
	"github.com/timutkin/otus-go/hw12_13_14_15_calendar/internal/storage"
)

var (
	ErrEventIDAlreadyExist = errors.New("event id already exist")
	ErrEventNotFound       = errors.New("event not found")
	EmptyEvent             = storage.Event{}
)

type Storage struct {
	db        *sqlx.DB
	dsn       string
	tableName string
}

func (s *Storage) Create(ctx context.Context, event storage.Event) error {
	sql := sq.Insert(s.tableName).
		Columns("id", "title", "date_time", "event_duration", "description", "user_id", "notification_time")
	_, err := sql.Values(event.ID, event.Title, event.DateTime, event.EventDuration, event.Description, event.UserID, event.NotificationTime). //nolint
																			RunWith(s.db).
																			ExecContext(ctx)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == UniqueViolation {
				return ErrEventIDAlreadyExist
			}
		}
		return err
	}
	return err
}

func (s *Storage) Update(ctx context.Context, newEvent storage.Event) error {
	sql := sq.Insert(s.tableName)
	if newEvent.UserID != nil {
		return storage.ErrUserIDShouldBeNil
	}
	if newEvent.Title != nil {
		sql = sql.Columns("title").Values(newEvent.Title)
	}
	if newEvent.DateTime != nil {
		sql = sql.Columns("date_time").Values(newEvent.DateTime)
	}
	if newEvent.EventDuration != nil {
		sql = sql.Columns("event_duration").Values(newEvent.EventDuration)
	}
	if newEvent.Description != nil {
		sql = sql.Columns("description").Values(newEvent.Description)
	}
	if newEvent.NotificationTime != nil {
		sql = sql.Columns("notification_time").Values(newEvent.NotificationTime)
	}
	_, err := sql.RunWith(s.db).
		ExecContext(ctx)
	return err
}

func (s *Storage) Delete(ctx context.Context, eventID uuid.UUID) error {
	_, err := sq.Delete(s.tableName).
		Where(sq.Eq{"id": eventID}).
		RunWith(s.db).
		ExecContext(ctx)
	return err
}

func (s *Storage) GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]storage.Event, error) {
	events := make([]storage.Event, 0)
	sql, args, err := sq.Select("*").From(s.tableName).Where(sq.Eq{"user_id": userID}).ToSql()
	if err != nil {
		return events, err
	}
	rows, err := s.db.NamedQueryContext(ctx, sql, args)
	if err != nil {
		return events, fmt.Errorf("error while executing select * from users where user_id = $1 : %w", err)
	}
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
	sql, args, err := sq.Select("*").From(s.tableName).Where(sq.Eq{"id": eventID}).ToSql()
	if err != nil {
		return EmptyEvent, err
	}
	rows, err := s.db.NamedQueryContext(ctx, sql, args)
	if err != nil {
		return EmptyEvent, fmt.Errorf("error while executing select * from users where id = $1 : %w", err)
	}
	if rows.Next() {
		var event storage.Event
		err := rows.StructScan(&event)
		if err != nil {
			return EmptyEvent, fmt.Errorf(ErrParsingToStructError, "storage.Event", err)
		}
	}
	return EmptyEvent, ErrEventNotFound
}

func New(cfg config.DBConf) *Storage {
	tables := cfg.DBTables
	return &Storage{
		dsn:       cfg.CollectDsn(),
		tableName: tables.Schema + "." + tables.Events,
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
