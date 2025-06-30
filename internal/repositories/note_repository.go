package repositories

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	appErrors "cu.ru/internal/errors"
	"cu.ru/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoteRepositoryInterface interface {
	GetNote(id int) (models.Note, error)
	CreateNote(note models.Note) (int, error)
	UpdateNote(id int, note models.Note) error
	DeleteNote(id int) error
	FindLike(text string) ([]models.Note, error)
}

type NoteRepositoryMemory struct {
	notes map[int]models.Note
	mu    sync.RWMutex
}

func NewNoteRepositoryMemory() *NoteRepositoryMemory {
	return &NoteRepositoryMemory{
		notes: make(map[int]models.Note),
	}
}

func (r *NoteRepositoryMemory) GetNote(id int) (models.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	note, exists := r.notes[id]
	if !exists {
		return models.Note{}, appErrors.ErrorNotFound
	}
	return note, nil
}

func (r *NoteRepositoryMemory) CreateNote(note models.Note) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := len(r.notes) + 1
	note.ID = id
	r.notes[id] = note
	return id, nil
}

func (r *NoteRepositoryMemory) UpdateNote(id int, note models.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existingNote, exists := r.notes[id]
	if !exists {
		return appErrors.ErrorNotFound
	}

	if note.Title != "" {
		existingNote.Title = note.Title
	}
	if note.Content != "" {
		existingNote.Content = note.Content
	}

	r.notes[id] = existingNote
	return nil
}

func (r *NoteRepositoryMemory) DeleteNote(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.notes[id]
	if !exists {
		return appErrors.ErrorNotFound
	}

	delete(r.notes, id)
	return nil
}

func (r *NoteRepositoryMemory) FindLike(text string) ([]models.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var notes []models.Note
	for _, note := range r.notes {
		if strings.Contains(note.Title, text) || strings.Contains(note.Content, text) {
			notes = append(notes, note)
		}
	}

	return notes, nil
}

type NoteRepositoryDb struct {
	pool *pgxpool.Pool
}

func NewNoteRepositoryDb(pool *pgxpool.Pool) *NoteRepositoryDb {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT NOT NULL
		);`)
	if err != nil {
		panic(err)
	}

	return &NoteRepositoryDb{
		pool: pool,
	}
}

func (r *NoteRepositoryDb) GetNote(id int) (models.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var note models.Note
	err := r.pool.QueryRow(ctx, "SELECT id, title, content FROM notes WHERE id = $1", id).Scan(&note.ID, &note.Title, &note.Content)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Note{}, appErrors.ErrorNotFound
		}
		return models.Note{}, err
	}
	return note, nil
}

func (r *NoteRepositoryDb) CreateNote(note models.Note) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id int
	err := r.pool.QueryRow(ctx, "INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING id", note.Title, note.Content).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *NoteRepositoryDb) UpdateNote(id int, note models.Note) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var args []interface{}
	var setClauses []string

	if note.Title != "" {
		setClauses = append(setClauses, "title = $"+strconv.Itoa(len(setClauses)+1))
		args = append(args, note.Title)
	}
	if note.Content != "" {
		setClauses = append(setClauses, "content = $"+strconv.Itoa(len(setClauses)+1))
		args = append(args, note.Content)
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	args = append(args, id)

	query := "UPDATE notes SET " + strings.Join(setClauses, ", ") + " WHERE id = $" + strconv.Itoa(len(args))

	_, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Println(query, args)
		}
		return err
	}
	return nil
}

func (r *NoteRepositoryDb) DeleteNote(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.pool.Exec(ctx, "DELETE FROM notes WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.ErrorNotFound
		}
		return err
	}
	return nil
}

func (r *NoteRepositoryDb) FindLike(text string) ([]models.Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, "SELECT id, title, content FROM notes WHERE title ILIKE $1 OR content ILIKE $1", "%"+text+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		if err := rows.Scan(&note.ID, &note.Title, &note.Content); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}
