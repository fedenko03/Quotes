package models

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Quote struct {
	ID      int
	Author  string
	Content string
	Created time.Time
}

type QuoteModel struct {
	DB *pgxpool.Pool
}

func (m *QuoteModel) Insert(author string, content string) (int, error) {
	row := m.DB.QueryRow(context.Background(),
		"INSERT INTO quotes  (author, content, created) VALUES ($1,$2, now()) RETURNING id;",
		author, content)
	var id uint64
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *QuoteModel) Get(id int) (*Quote, error) {
	s := &Quote{}
	stmt := `SELECT * FROM quotes 
WHERE id = $1`
	err := m.DB.QueryRow(context.Background(), stmt, id).Scan(&s.ID, &s.Author, &s.Content, &s.Created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *QuoteModel) Latest() ([]*Quote, error) {
	stmt := `SELECT id, author, content, created FROM quotes
ORDER BY id DESC LIMIT 10`
	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	quotes := []*Quote{}
	for rows.Next() {
		s := &Quote{}
		err = rows.Scan(&s.ID, &s.Author, &s.Content, &s.Created)
		if err != nil {
			return nil, err
		}
		quotes = append(quotes, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return quotes, nil
}
