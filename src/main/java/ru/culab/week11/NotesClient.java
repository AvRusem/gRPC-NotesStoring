package ru.culab.week11;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import ru.culab.week11.NotesServiceGrpc.NotesServiceBlockingStub;

public class NotesClient implements NotesConnectable {
    private ManagedChannel channel = null;
    private NotesServiceBlockingStub stub = null;

    @Override
    public void connectToServer(String hostName, int port) throws Exception {
        channel = ManagedChannelBuilder.forAddress(hostName, port)
                .usePlaintext()
                .build();
        stub = NotesServiceGrpc.newBlockingStub(channel);
    }

    @Override
    public void close() {
        channel.shutdown();
    }

    @Override
    public long createNote(String title, String content) throws Exception {
        NoteRequest noteReq = NoteRequest.newBuilder().setTitle(title).setContent(content).build();
        Note noteResp = stub.createNote(noteReq);

        return noteResp.getId();
    }

    @Override
    public String[] getNoteTitleAndContent(long id) throws Exception {
        IdRequest idReq = IdRequest.newBuilder().setId(id).build();
        Note noteResp = stub.getNote(idReq);

        return new String[] {noteResp.getTitle(), noteResp.getContent()};
    }

    @Override
    public void updateNote(long id, String title, String content) throws Exception {
        Note noteReq = Note.newBuilder().setId(id).setTitle(title).setContent(content).build();
        stub.updateNote(noteReq);
    }

    @Override
    public void deleteNote(long id) throws Exception {
        IdRequest idReq = IdRequest.newBuilder().setId(id).build();
        stub.deleteNote(idReq);
    }

    @Override
    public long[] searchNotes(String pattern) throws Exception {
        SearchRequest searchReq = SearchRequest.newBuilder().setPattern(pattern).build();
        Notes notes = stub.searchNotes(searchReq);

        return notes.getNotesList().stream()
            .mapToLong(Note::getId)
            .toArray();
    }
}