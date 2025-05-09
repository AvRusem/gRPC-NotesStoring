package services

import (
	"cu.ru/internal/models"
	"cu.ru/internal/repositories"
)

type NoteService struct {
	repository repositories.NoteRepositoryInterface
}

func NewNoteService(repository repositories.NoteRepositoryInterface) *NoteService {
	return &NoteService{
		repository: repository,
	}
}

func (s *NoteService) GetNote(id int) (models.Note, error) {
	note, err := s.repository.GetNote(id)
	if err != nil {
		return models.Note{}, err
	}
	return note, nil
}

func (s *NoteService) CreateNote(note models.Note) (int, error) {
	id, err := s.repository.CreateNote(note)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *NoteService) UpdateNote(note models.Note) error {
	err := s.repository.UpdateNote(note.ID, note)
	if err != nil {
		return err
	}
	return nil
}

func (s *NoteService) DeleteNote(id int) error {
	err := s.repository.DeleteNote(id)
	if err != nil {
		return err
	}
	return nil
}

func (s *NoteService) FindLike(text string) ([]models.Note, error) {
	notes, err := s.repository.FindLike(text)
	if err != nil {
		return nil, err
	}
	return notes, nil
}
