package mappers

import (
	"cu.ru/internal/models"
	"cu.ru/pb"
)

func NoteFromProto(note *pb.NoteRequest) models.Note {
	return models.Note{
		Title:   note.GetTitle(),
		Content: note.GetContent(),
	}
}

func NoteToProto(note models.Note) *pb.Note {
	return &pb.Note{
		Id:      int64(note.ID),
		Title:   note.Title,
		Content: note.Content,
	}
}

func NoteFromProtoNote(note *pb.Note) models.Note {
	return models.Note{
		ID:      int(note.GetId()),
		Title:   note.GetTitle(),
		Content: note.GetContent(),
	}
}
