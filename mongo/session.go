package mongo

import (
	"context"

	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Session struct {
	session mongodb.Session
}

func NewSession(db *mongodb.Database) (*Session, error) {
	s, err := db.Client().StartSession()
	if err != nil {
		return nil, err
	}
	return &Session{session: s}, nil
}

func (s *Session) RunWithTransaction(fn func(ctx context.Context) error, opt ...*options.TransactionOptions) error {
	defer s.session.EndSession(context.Background())
	err := s.session.StartTransaction(opt...)
	if err != nil {
		return err
	}
	return mongodb.WithSession(context.Background(), s.session, func(sc mongodb.SessionContext) error {
		err := fn(sc)
		if err != nil {
			s.session.AbortTransaction(context.Background())
			return err
		} else {
			return s.session.CommitTransaction(context.Background())
		}
	})
}
