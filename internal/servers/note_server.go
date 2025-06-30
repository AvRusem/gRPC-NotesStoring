package servers

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	appErrors "cu.ru/internal/errors"
	"cu.ru/internal/mappers"
	"cu.ru/internal/repositories"
	"cu.ru/internal/services"
	pb "cu.ru/pb"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	InternalServerError = "internal server error: %v"
	InvalidNote         = "invalid note: %v"
	NoteNotFound        = "note not found: %v"
	PatternEmpty        = "pattern cannot be empty"
)

type notesServer struct {
	pb.UnimplementedNotesServiceServer
	service  *services.NoteService
	validate *validator.Validate
}

func NewNotesServer(service *services.NoteService) *notesServer {
	return &notesServer{
		service:  service,
		validate: validator.New(),
	}
}

func (s *notesServer) CreateNote(ctx context.Context, req *pb.NoteRequest) (*pb.Note, error) {
	note := mappers.NoteFromProto(req)
	if err := s.validate.Struct(note); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, InvalidNote, err)
	}

	id, err := s.service.CreateNote(note)
	if err != nil {
		return nil, status.Errorf(codes.Internal, InternalServerError, err)
	}

	note.ID = id
	return mappers.NoteToProto(note), nil
}

func (s *notesServer) GetNote(ctx context.Context, req *pb.IdRequest) (*pb.Note, error) {
	note, err := s.service.GetNote(int(req.GetId()))
	if err != nil {
		if err == appErrors.ErrorNotFound {
			return nil, status.Errorf(codes.NotFound, NoteNotFound, err)
		}
		return nil, status.Errorf(codes.Internal, InternalServerError, err)
	}

	return mappers.NoteToProto(note), nil
}

func (s *notesServer) UpdateNote(ctx context.Context, req *pb.Note) (*pb.Empty, error) {
	note := mappers.NoteFromProtoNote(req)
	if err := s.validate.Struct(note); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, InvalidNote, err)
	}

	err := s.service.UpdateNote(note)
	if err != nil {
		if err == appErrors.ErrorNotFound {
			return nil, status.Errorf(codes.NotFound, NoteNotFound, err)
		}
		return nil, status.Errorf(codes.Internal, InternalServerError, err)
	}

	return &pb.Empty{}, nil
}

func (s *notesServer) DeleteNote(ctx context.Context, req *pb.IdRequest) (*pb.Empty, error) {
	err := s.service.DeleteNote(int(req.GetId()))
	if err != nil {
		if err == appErrors.ErrorNotFound {
			return nil, status.Errorf(codes.NotFound, NoteNotFound, err)
		}
		return nil, status.Errorf(codes.Internal, InternalServerError, err)
	}

	return &pb.Empty{}, nil
}

func (s *notesServer) SearchNotes(ctx context.Context, req *pb.SearchRequest) (*pb.Notes, error) {
	pattern := req.GetPattern()
	if pattern == "" {
		return nil, status.Errorf(codes.InvalidArgument, PatternEmpty)
	}

	notes, err := s.service.FindLike(pattern)
	if err != nil {
		return nil, status.Errorf(codes.Internal, InternalServerError, err)
	}

	protoNotes := make([]*pb.Note, len(notes))
	for i, note := range notes {
		protoNotes[i] = mappers.NoteToProto(note)
	}

	return &pb.Notes{
		Notes: protoNotes,
	}, nil
}

func getService(dsn string) *services.NoteService {
	var repo repositories.NoteRepositoryInterface
	if dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			panic(err)
		}

		repo = repositories.NewNoteRepositoryDb(pool)
	} else {
		repo = repositories.NewNoteRepositoryMemory()
	}
	return services.NewNoteService(repo)
}

func buildServer(addr, dsn string) (*grpc.Server, net.Listener) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNotesServiceServer(grpcServer, NewNotesServer(getService(dsn)))

	return grpcServer, listener
}

func run(server *grpc.Server, listener net.Listener) {
	log.Printf("gRPC-server running on %v...", listener.Addr().String())

	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

func StartServer(addr, dsn string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	server, listener := buildServer(addr, dsn)

	go run(server, listener)

	<-stop
	log.Println("Server is shutting down...")

	server.GracefulStop()

	log.Println("Server gracefully stopped")
}
