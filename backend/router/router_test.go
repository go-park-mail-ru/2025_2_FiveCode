package router

import (
	"backend/initialize"
	"backend/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

// minimal fakes to satisfy deliveries interfaces
type fakeAuth struct{}
func (f *fakeAuth) Login(w http.ResponseWriter, r *http.Request)      {}
func (f *fakeAuth) Logout(w http.ResponseWriter, r *http.Request)     {}
func (f *fakeAuth) Register(w http.ResponseWriter, r *http.Request)   {}

type fakeUser struct{}
func (f *fakeUser) GetProfile(w http.ResponseWriter, r *http.Request)       {}
func (f *fakeUser) GetProfileBySession(w http.ResponseWriter, r *http.Request) {}
func (f *fakeUser) UpdateProfile(w http.ResponseWriter, r *http.Request)    {}

type fakeNotes struct{}
func (f *fakeNotes) GetAllNotes(w http.ResponseWriter, r *http.Request) {}
func (f *fakeNotes) CreateNote(w http.ResponseWriter, r *http.Request)  {}
func (f *fakeNotes) GetNoteById(w http.ResponseWriter, r *http.Request) {}
func (f *fakeNotes) UpdateNote(w http.ResponseWriter, r *http.Request)  {}
func (f *fakeNotes) DeleteNote(w http.ResponseWriter, r *http.Request)  {}
func (f *fakeNotes) AddFavorite(w http.ResponseWriter, r *http.Request) {}
func (f *fakeNotes) RemoveFavorite(w http.ResponseWriter, r *http.Request) {}

type fakeBlocks struct{}
func (f *fakeBlocks) CreateBlock(w http.ResponseWriter, r *http.Request)       {}
func (f *fakeBlocks) GetBlocks(w http.ResponseWriter, r *http.Request)         {}
func (f *fakeBlocks) GetBlock(w http.ResponseWriter, r *http.Request)          {}
func (f *fakeBlocks) UpdateBlock(w http.ResponseWriter, r *http.Request)       {}
func (f *fakeBlocks) DeleteBlock(w http.ResponseWriter, r *http.Request)       {}
func (f *fakeBlocks) UpdateBlockPosition(w http.ResponseWriter, r *http.Request) {}

type fakeFile struct{}
func (f *fakeFile) UploadFile(w http.ResponseWriter, r *http.Request) {}
func (f *fakeFile) GetFile(w http.ResponseWriter, r *http.Request)    {}
func (f *fakeFile) DeleteFile(w http.ResponseWriter, r *http.Request) {}

// fake redis implementing only GetUserIDBySession used by AuthMiddleware
type fakeRedis struct{
    userID uint64
    err    error
}
func (f *fakeRedis) GetUserIDBySession(_ any, _ string) (uint64, error) { return f.userID, f.err }

type fakeStore struct{
    Redis *fakeRedis
}

func TestNewRouter_BasicRoutes(t *testing.T) {
    deliveries := &initialize.Deliveries{
        AuthDelivery:   &fakeAuth{},
        UserDelivery:   &fakeUser{},
        NotesDelivery:  &fakeNotes{},
        BlocksDelivery: &fakeBlocks{},
        FileDelivery:   &fakeFile{},
    }

    s := &store.Store{Redis: nil}

    handler := NewRouter(s, deliveries)

    // basic smoke: GET /api/notes should route (unauthenticated will return 400 from AuthMiddleware)
    req := httptest.NewRequest("GET", "/api/notes", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)

    // AuthMiddleware without a session cookie should produce 400 (no session cookie)
    if w.Result().StatusCode != http.StatusBadRequest {
        t.Fatalf("expected 400 got %d", w.Result().StatusCode)
    }

    // Test swagger path exists
    req2 := httptest.NewRequest("GET", "/swagger/index.html", nil)
    w2 := httptest.NewRecorder()
    handler.ServeHTTP(w2, req2)
    if w2.Result().StatusCode == http.StatusNotFound {
        t.Fatalf("swagger handler not found, got 404")
    }
}
