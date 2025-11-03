package store

import (
	"backend/config"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	Mu       sync.RWMutex
	Minio    *MinioStorage
	Postgres *PostgresDB
	Redis    *RedisDB

	Users        map[uint64]*models.User
	UsersByEmail map[string]uint64
	Notes        map[uint64]*models.Note
	Blocks       map[uint64]*models.Block
	BlockTexts   map[uint64]*models.BlockText
	BlockFormats map[uint64][]models.BlockTextFormat
	Files        map[uint64]*models.File
	sessions     map[string]uint64

	nextUserID      uint64
	nextNoteID      uint64
	nextBlockID     uint64
	nextBlockTextID uint64
	nextFormatID    uint64
	nextFileID      uint64
}

func (s *Store) InitRedis(conf *config.Config) error {
	rdb, err := NewRedisDB(
		conf.Storages.Redis.Host,
		conf.Storages.Redis.Port,
		conf.Storages.Redis.Password,
		conf.Storages.Redis.DB,
	)
	if err != nil {
		return fmt.Errorf("failed to init redis: %w", err)
	}

	s.Redis = rdb
	return nil
}

func (s *Store) InitPostgres(conf *config.Config) error {
	pg, err := NewPostgresDB(
		conf.Storages.Db.Host,
		conf.Storages.Db.Port,
		conf.Storages.Db.User,
		conf.Storages.Db.Password,
		conf.Storages.Db.DBName,
		conf.Storages.Db.SSLMode,
	)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}

	s.Postgres = pg
	return nil
}

func (s *Store) InitMinioStorage(conf *config.Config) error {
	minioStorage, err := NewMinioStorage(
		conf.Storages.Minio.Endpoint,
		conf.Storages.Minio.AccessKey,
		conf.Storages.Minio.SecretKey,
		conf.Storages.Minio.Secure,
	)
	if err != nil {
		return errors.New("Error to init Minio storage: " + err.Error())
	}

	s.Minio = minioStorage
	return nil
}

func (s *Store) InitFillStore() error {
	ctx := context.Background()
	email := "user@example.com"
	password := "password"

	var userID uint64
	var exists bool
	checkQuery := `SELECT id FROM "user" WHERE email = $1`
	err := s.Postgres.DB.QueryRowContext(ctx, checkQuery, email).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		username := strings.Split(email, "@")[0]
		insertQuery := `
            INSERT INTO "user" (email, password_hash, username)
            VALUES ($1, $2, $3)
            RETURNING id
        `
		err = s.Postgres.DB.QueryRowContext(ctx, insertQuery, email, string(hashedPassword), username).Scan(&userID)
		if err != nil {
			return fmt.Errorf("failed to create user in PostgreSQL: %w", err)
		}
		exists = false
	} else if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	} else {
		exists = true
	}

	user, err := s.CreateUser(email, password)
	if err != nil && !errors.Is(err, namederrors.ErrUserExists) {
		return fmt.Errorf("failed to create user in memory: %w", err)
	}
	if err == nil {
		s.Mu.Lock()
		delete(s.Users, user.ID)
		delete(s.UsersByEmail, email)

		user.ID = userID
		s.Users[userID] = user
		s.UsersByEmail[email] = userID
		s.Mu.Unlock()

	} else {
	}

	if !exists {
		notes := []struct {
			Title     string
			IsShared  bool
			CreatedAt time.Time
			UpdatedAt time.Time
		}{
			{
				Title:     "University Lectures",
				IsShared:  false,
				CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
			},
			{
				Title:     "Project Ideas",
				IsShared:  true,
				CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
			},
			{
				Title:     "Shopping List",
				IsShared:  false,
				CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-6 * time.Hour),
			},
			{
				Title:     "Random Note",
				IsShared:  false,
				CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-8 * 24 * time.Hour),
			},
		}

		for _, note := range notes {
			insertNoteQuery := `
                INSERT INTO note (owner_id, title, is_archived, is_shared, created_at, updated_at)
                VALUES ($1, $2, $3, $4, $5, $6)
                RETURNING id
            `
			var noteID uint64
			err = s.Postgres.DB.QueryRowContext(
				ctx,
				insertNoteQuery,
				userID,
				note.Title,
				false,
				note.IsShared,
				note.CreatedAt,
				note.UpdatedAt,
			).Scan(&noteID)

			if err != nil {
				return fmt.Errorf("failed to create note '%s': %w", note.Title, err)
			}

		}
	} else {
	}

	log.Info().Msg("InitFillStore completed successfully")
	return nil
}

func NewStore() *Store {
	return &Store{
		Users:        make(map[uint64]*models.User),
		UsersByEmail: make(map[string]uint64),
		Notes:        make(map[uint64]*models.Note),
		Blocks:       make(map[uint64]*models.Block),
		BlockTexts:   make(map[uint64]*models.BlockText),
		BlockFormats: make(map[uint64][]models.BlockTextFormat),
		Files:        make(map[uint64]*models.File),
		sessions:     make(map[string]uint64),

		nextUserID:      1,
		nextNoteID:      1,
		nextBlockID:     1,
		nextBlockTextID: 1,
		nextFormatID:    1,
		nextFileID:      1,
	}
}

func (s *Store) CreateDefaultNotes(userID uint64) {
	notes := []*models.Note{
		{
			ID:           s.nextNoteID,
			OwnerID:      userID,
			ParentNoteID: nil,
			Title:        "University Lectures",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-5 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           s.nextNoteID + 1,
			OwnerID:      userID,
			ParentNoteID: nil,
			Title:        "Project Ideas",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     true,
			CreatedAt:    time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-2 * 24 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           s.nextNoteID + 2,
			OwnerID:      userID,
			ParentNoteID: nil,
			Title:        "Shopping List",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-6 * time.Hour),
			DeletedAt:    nil,
		},
		{
			ID:           s.nextNoteID + 3,
			OwnerID:      userID,
			ParentNoteID: nil,
			Title:        "Random Note",
			IconFileID:   nil,
			IsArchived:   false,
			IsShared:     false,
			CreatedAt:    time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-8 * 24 * time.Hour),
			DeletedAt:    nil,
		},
	}
	for _, note := range notes {
		s.Notes[note.ID] = note
	}
	s.nextNoteID += 4
}

func (s *Store) CreateUser(email, password string) (*models.User, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if _, ok := s.UsersByEmail[email]; ok {
		return nil, namederrors.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Cannot hash password:" + err.Error())
	}

	user := &models.User{
		ID:        s.nextUserID,
		Email:     email,
		Username:  fmt.Sprintf("user_%d", s.nextUserID),
		Password:  string(hashedPassword),
		CreatedAt: time.Now().UTC(),
	}
	s.Users[user.ID] = user
	s.UsersByEmail[email] = user.ID
	s.CreateDefaultNotes(user.ID)
	s.nextUserID++

	return user, nil
}

func (s *Store) AuthenticateUser(email, password string) (*models.User, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.UsersByEmail[email]
	if !ok {
		return nil, namederrors.ErrInvalidEmailOrPassword
	}
	user := s.Users[userID]

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, namederrors.ErrInvalidEmailOrPassword
	}

	return user, nil
}

func (s *Store) CreateSession(userID uint64) string {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	sessionID := uuid.NewString()
	s.sessions[sessionID] = userID

	return sessionID
}

func (s *Store) DeleteSession(sessionID string) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	delete(s.sessions, sessionID)
}

func (s *Store) GetUserBySession(sessionID string) (*models.User, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		log.Info().Str("session_id", sessionID).Msg("session not found")
		return nil, false
	}
	user, ok := s.Users[userID]

	return user, ok
}

func (s *Store) GetUserIDBySession(sessionID string) (uint64, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	userID, ok := s.sessions[sessionID]
	if !ok {
		log.Info().Str("session_id", sessionID).Msg("session not found")
		return 0, false
	}

	return userID, true
}

func (s *Store) ListNotes(ownerID uint64) []models.Note {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	result := make([]models.Note, 0)
	for _, note := range s.Notes {
		if note.OwnerID == ownerID && note.DeletedAt == nil {
			result = append(result, *note)
		}
	}

	return result
}

func (s *Store) CreateNote(userID uint64) (*models.Note, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	result := &models.Note{
		ID:        s.nextNoteID,
		OwnerID:   userID,
		Title:     "",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	s.Notes[result.ID] = result
	s.nextNoteID++
	return result, nil
}

func (s *Store) GetNoteById(noteID uint64) (*models.Note, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	result := s.Notes[noteID]
	if result == nil {
		return nil, namederrors.ErrNotFound
	}

	return result, nil
}

func (s *Store) DeleteNote(noteID uint64) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	delete(s.Notes, noteID)

	return nil
}

func (s *Store) UpdateNote(noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.Notes[noteID] == nil {
		return nil, namederrors.ErrNotFound
	}

	if title != nil {
		s.Notes[noteID].Title = *title
	}

	if isArchived != nil {
		s.Notes[noteID].IsArchived = *isArchived
	}
	s.Notes[noteID].UpdatedAt = time.Now().UTC()

	return s.Notes[noteID], nil
}

func (s *Store) CreateBlock(noteID uint64, blockType models.BlockType, beforeBlockID *uint64) (*models.Block, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if _, ok := s.Notes[noteID]; !ok {
		return nil, namederrors.ErrNotFound
	}

	position, err := s.calculatePosition(noteID, beforeBlockID, 0)
	if err != nil {
		return nil, err
	}

	block := &models.Block{
		ID:        s.nextBlockID,
		NoteID:    noteID,
		Type:      blockType,
		Position:  position,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	s.Blocks[block.ID] = block
	s.nextBlockID++

	if blockType == models.BlockTypeText {
		blockText := &models.BlockText{
			ID:        s.nextBlockTextID,
			BlockID:   block.ID,
			Text:      "",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		s.BlockTexts[blockText.ID] = blockText
		s.nextBlockTextID++
	}

	return block, nil
}

func (s *Store) calculatePosition(noteID uint64, beforeBlockID *uint64, excludeBlockID uint64) (float64, error) {
	noteBlocks := make([]*models.Block, 0)
	for _, block := range s.Blocks {
		if block.NoteID == noteID && block.ID != excludeBlockID {
			noteBlocks = append(noteBlocks, block)
		}
	}

	if len(noteBlocks) == 0 {
		return 1.0, nil
	}

	if beforeBlockID == nil {
		maxPos := noteBlocks[0].Position
		for _, b := range noteBlocks {
			if b.Position > maxPos {
				maxPos = b.Position
			}
		}
		return maxPos + 1.0, nil
	}

	beforeBlock, ok := s.Blocks[*beforeBlockID]
	if !ok {
		return 0, errors.New("before_block not found")
	}
	if beforeBlock.NoteID != noteID {
		return 0, errors.New("before_block belongs to different note")
	}

	var prevBlock *models.Block
	for _, b := range noteBlocks {
		if b.Position < beforeBlock.Position {
			if prevBlock == nil || b.Position > prevBlock.Position {
				prevBlock = b
			}
		}
	}

	if prevBlock == nil {
		return beforeBlock.Position / 2.0, nil
	}

	return (prevBlock.Position + beforeBlock.Position) / 2.0, nil
}

func (s *Store) GetBlocksByNoteID(noteID uint64) ([]models.BlockWithContent, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	blocks := s.getBlocksByNoteIDUnsafe(noteID)

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Position < blocks[j].Position
	})

	result := make([]models.BlockWithContent, 0, len(blocks))
	for _, block := range blocks {
		blockWithContent := models.BlockWithContent{
			Block: *block,
		}

		if block.Type == models.BlockTypeText {
			blockText := s.getBlockTextByBlockIDUnsafe(block.ID)
			if blockText != nil {
				blockWithContent.Text = blockText.Text
				blockWithContent.Formats = s.BlockFormats[blockText.ID]
			}
		}

		result = append(result, blockWithContent)
	}

	return result, nil
}

func (s *Store) getBlocksByNoteIDUnsafe(noteID uint64) []*models.Block {
	blocks := make([]*models.Block, 0)
	for _, block := range s.Blocks {
		if block.NoteID == noteID {
			blocks = append(blocks, block)
		}
	}
	return blocks
}

func (s *Store) getBlockTextByBlockIDUnsafe(blockID uint64) *models.BlockText {
	for _, bt := range s.BlockTexts {
		if bt.BlockID == blockID {
			return bt
		}
	}
	return nil
}

func (s *Store) GetBlockByID(blockID uint64) (*models.BlockWithContent, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	block, ok := s.Blocks[blockID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	result := &models.BlockWithContent{
		Block: *block,
	}

	if block.Type == models.BlockTypeText {
		blockText := s.getBlockTextByBlockIDUnsafe(block.ID)
		if blockText != nil {
			result.Text = blockText.Text
			result.Formats = s.BlockFormats[blockText.ID]
		}
	}

	return result, nil
}

func (s *Store) UpdateBlockText(blockID uint64, text string, formats []models.BlockTextFormat) (*models.BlockWithContent, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	block, ok := s.Blocks[blockID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	if block.Type != models.BlockTypeText {
		return nil, errors.New("block is not text type")
	}

	block.UpdatedAt = time.Now().UTC()

	blockText := s.getBlockTextByBlockIDUnsafe(blockID)
	if blockText == nil {
		blockText = &models.BlockText{
			ID:        s.nextBlockTextID,
			BlockID:   blockID,
			Text:      text,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		s.BlockTexts[blockText.ID] = blockText
		s.nextBlockTextID++
	} else {
		blockText.Text = text
		blockText.UpdatedAt = time.Now().UTC()
	}

	optimizedFormats := s.optimizeFormats(text, formats)

	for i := range optimizedFormats {
		optimizedFormats[i].ID = s.nextFormatID
		optimizedFormats[i].BlockTextID = blockText.ID
		optimizedFormats[i].CreatedAt = time.Now().UTC()
		optimizedFormats[i].UpdatedAt = time.Now().UTC()
		s.nextFormatID++
	}

	s.BlockFormats[blockText.ID] = optimizedFormats

	return &models.BlockWithContent{
		Block:   *block,
		Text:    blockText.Text,
		Formats: optimizedFormats,
	}, nil
}

func (s *Store) optimizeFormats(text string, formats []models.BlockTextFormat) []models.BlockTextFormat {
	if len(formats) == 0 {
		return []models.BlockTextFormat{}
	}

	textLen := len(text)

	validFormats := make([]models.BlockTextFormat, 0)
	for _, f := range formats {
		if f.StartOffset >= 0 && f.EndOffset <= textLen && f.StartOffset < f.EndOffset {
			if !isDefaultFormat(f) {
				validFormats = append(validFormats, f)
			}
		}
	}

	if len(validFormats) == 0 {
		return []models.BlockTextFormat{}
	}

	type event struct {
		offset int
		format models.BlockTextFormat
		isEnd  bool
	}

	events := make([]event, 0)
	for _, f := range validFormats {
		events = append(events, event{offset: f.StartOffset, format: f, isEnd: false})
		events = append(events, event{offset: f.EndOffset, format: f, isEnd: true})
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].offset == events[j].offset {
			return !events[i].isEnd && events[j].isEnd
		}
		return events[i].offset < events[j].offset
	})

	activeFormats := make(map[int]models.BlockTextFormat)
	result := make([]models.BlockTextFormat, 0)
	lastOffset := 0
	formatIndex := 0

	for _, ev := range events {
		if len(activeFormats) > 0 && ev.offset > lastOffset {
			merged := mergeFormats(activeFormats)
			merged.StartOffset = lastOffset
			merged.EndOffset = ev.offset
			result = append(result, merged)
		}

		lastOffset = ev.offset

		if ev.isEnd {
			for idx, f := range activeFormats {
				if formatsEqual(f, ev.format) {
					delete(activeFormats, idx)
					break
				}
			}
		} else {
			activeFormats[formatIndex] = ev.format
			formatIndex++
		}
	}

	if len(result) == 0 {
		return result
	}

	merged := make([]models.BlockTextFormat, 0)
	current := result[0]

	for i := 1; i < len(result); i++ {
		if current.EndOffset == result[i].StartOffset && stylesEqual(current, result[i]) {
			current.EndOffset = result[i].EndOffset
		} else {
			merged = append(merged, current)
			current = result[i]
		}
	}
	merged = append(merged, current)

	return merged
}

func isDefaultFormat(f models.BlockTextFormat) bool {
	return !f.Bold && !f.Italic && !f.Underline && !f.Strikethrough &&
		f.Link == nil && f.Font == models.FontInter && f.Size == 12
}

func formatsEqual(f1, f2 models.BlockTextFormat) bool {
	return f1.StartOffset == f2.StartOffset &&
		f1.EndOffset == f2.EndOffset &&
		f1.Bold == f2.Bold &&
		f1.Italic == f2.Italic &&
		f1.Underline == f2.Underline &&
		f1.Strikethrough == f2.Strikethrough &&
		((f1.Link == nil && f2.Link == nil) || (f1.Link != nil && f2.Link != nil && *f1.Link == *f2.Link)) &&
		f1.Font == f2.Font &&
		f1.Size == f2.Size
}

func stylesEqual(f1, f2 models.BlockTextFormat) bool {
	return f1.Bold == f2.Bold &&
		f1.Italic == f2.Italic &&
		f1.Underline == f2.Underline &&
		f1.Strikethrough == f2.Strikethrough &&
		((f1.Link == nil && f2.Link == nil) || (f1.Link != nil && f2.Link != nil && *f1.Link == *f2.Link)) &&
		f1.Font == f2.Font &&
		f1.Size == f2.Size
}

func mergeFormats(formats map[int]models.BlockTextFormat) models.BlockTextFormat {
	result := models.BlockTextFormat{
		Font: models.FontInter,
		Size: 12,
	}

	for _, f := range formats {
		if f.Bold {
			result.Bold = true
		}
		if f.Italic {
			result.Italic = true
		}
		if f.Underline {
			result.Underline = true
		}
		if f.Strikethrough {
			result.Strikethrough = true
		}
		if f.Link != nil {
			result.Link = f.Link
		}
		result.Font = f.Font
		result.Size = f.Size
	}

	return result
}

func (s *Store) DeleteBlock(blockID uint64) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	block, ok := s.Blocks[blockID]
	if !ok {
		return namederrors.ErrNotFound
	}

	if block.Type == models.BlockTypeText {
		blockText := s.getBlockTextByBlockIDUnsafe(blockID)
		if blockText != nil {
			delete(s.BlockFormats, blockText.ID)
			delete(s.BlockTexts, blockText.ID)
		}
	}

	delete(s.Blocks, blockID)

	return nil
}

func (s *Store) UpdateBlockPosition(blockID uint64, beforeBlockID *uint64) (*models.Block, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	block, ok := s.Blocks[blockID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	position, err := s.calculatePosition(block.NoteID, beforeBlockID, blockID)
	if err != nil {
		return nil, errors.New("Error calculating position: " + err.Error())
	}

	block.Position = position
	block.UpdatedAt = time.Now().UTC()

	return block, nil
}

func (s *Store) GetBlockNoteID(blockID uint64) (uint64, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	block, ok := s.Blocks[blockID]
	if !ok {
		return 0, namederrors.ErrNotFound
	}

	return block.NoteID, nil
}

func (s *Store) UpdateUserProfile(userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	user, ok := s.Users[userID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	if username != nil {
		user.Username = *username
	}
	if password != nil {
		user.Password = *password
	}
	if avatarFileID != nil {
		user.AvatarFileID = avatarFileID
	}

	now := time.Now().UTC()
	user.UpdatedAt = &now

	return user, nil
}

func (s *Store) GetUserByID(userID uint64) (*models.User, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	user, ok := s.Users[userID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	return user, nil
}

func (s *Store) SaveFile(file *models.File) (*models.File, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	file.ID = s.nextFileID
	s.Files[file.ID] = file
	s.nextFileID++

	return file, nil
}

func (s *Store) GetFileByID(fileID uint64) (*models.File, error) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	file, ok := s.Files[fileID]
	if !ok {
		return nil, namederrors.ErrNotFound
	}

	return file, nil
}

func (s *Store) UpdateFile(file *models.File) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, ok := s.Files[file.ID]
	if !ok {
		return namederrors.ErrNotFound
	}

	s.Files[file.ID] = file
	return nil
}

func (s *Store) DeleteFile(fileID uint64) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	_, ok := s.Files[fileID]
	if !ok {
		return namederrors.ErrNotFound
	}

	delete(s.Files, fileID)
	return nil
}

func (s *Store) UploadFileToMinIO(ctx context.Context, filename string, file io.Reader, size int64, contentType string) (string, error) {
	if s.Minio == nil {
		return "", errors.New("minio storage not initialized")
	}

	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)
	bucketName := "notes-app"

	client := s.Minio.GetClient()
	_, err := client.PutObject(ctx, bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	endpoint := client.EndpointURL()
	scheme := endpoint.Scheme
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint.Host, bucketName, objectName)
	return url, nil
}
