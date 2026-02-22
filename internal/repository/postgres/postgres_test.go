package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/MaxRadzey/shortener/internal/models"
	"github.com/MaxRadzey/shortener/internal/repository"
	"github.com/MaxRadzey/shortener/internal/repository/postgres"
	"github.com/MaxRadzey/shortener/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testDSN      string
	testDB       *pgxpool.Pool
	dbAvailable  bool
	migrationsOK bool
)

const defaultTestDSN = "postgres://shortener:shortener@localhost:5432/shortener?sslmode=disable"

func TestMain(m *testing.M) {
	testDSN = os.Getenv("TEST_DATABASE_DSN")
	if testDSN == "" {
		testDSN = defaultTestDSN
	}

	// Выполняем миграции
	migrationsPath := filepath.Join("..", "..", "..", "migrations")
	absMigrationsPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		os.Stderr.WriteString("failed to get migrations path: " + err.Error() + "\n")
		os.Stderr.WriteString("Tests will be skipped\n")
		migrationsOK = false
		os.Exit(m.Run())
		return
	}

	if err = storage.RunMigrations(testDSN, absMigrationsPath); err != nil {
		os.Stderr.WriteString("migrations: " + err.Error() + "\n")
		os.Stderr.WriteString("Tests will be skipped\n")
		migrationsOK = false
		os.Exit(m.Run())
		return
	}
	migrationsOK = true

	// Создаем подключение к БД
	testDB, err = pgxpool.New(context.Background(), testDSN)
	if err != nil {
		os.Stderr.WriteString("open db: " + err.Error() + "\n")
		os.Stderr.WriteString("Tests will be skipped\n")
		dbAvailable = false
		os.Exit(m.Run())
		return
	}
	defer func() {
		if testDB != nil {
			testDB.Close()
		}
	}()

	// Проверяем подключение
	if err := testDB.Ping(context.Background()); err != nil {
		os.Stderr.WriteString("ping db: " + err.Error() + "\n")
		os.Stderr.WriteString("Tests will be skipped\n")
		dbAvailable = false
		os.Exit(m.Run())
		return
	}

	dbAvailable = true
	os.Exit(m.Run())
}

// skipIfDBUnavailable пропускает тест, если БД недоступна или миграции не выполнены.
func skipIfDBUnavailable(t *testing.T) {
	t.Helper()
	if !dbAvailable || !migrationsOK || testDB == nil {
		t.Skip("Skipping test: database connection unavailable or migrations failed")
	}
}

// setupDB очищает таблицу urls и возвращает репозиторий для тестов.
func setupDB(t *testing.T) *postgres.PostgresRepository {
	t.Helper()
	skipIfDBUnavailable(t)
	ctx := context.Background()

	// Очищаем таблицу перед каждым тестом
	_, err := testDB.Exec(ctx, "TRUNCATE TABLE urls RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to truncate table")

	// Создаем репозиторий
	repo, err := postgres.NewPostgresRepository(testDB)
	require.NoError(t, err, "Failed to create repository")

	return repo
}

func TestPostgresRepository_Get(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		// Создаем запись
		entry := models.URLEntry{
			ShortPath: "abc123",
			FullURL:   "https://example.com",
			UserID:    uuid.New().String(),
		}
		err := repo.Create(entry)
		require.NoError(t, err)

		// Получаем запись
		originalURL, err := repo.Get("abc123")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", originalURL)
	})

	t.Run("not found", func(t *testing.T) {
		originalURL, err := repo.Get("nonexistent")
		assert.Error(t, err)
		assert.Empty(t, originalURL)

		var notFoundErr *repository.ErrNotFound
		assert.ErrorAs(t, err, &notFoundErr)
		assert.Equal(t, "nonexistent", notFoundErr.ShortPath)
	})

	t.Run("gone - deleted URL", func(t *testing.T) {
		// Создаем запись
		entry := models.URLEntry{
			ShortPath: "deleted123",
			FullURL:   "https://deleted.com",
			UserID:    uuid.New().String(),
		}
		err := repo.Create(entry)
		require.NoError(t, err)

		// Помечаем как удаленную напрямую через SQL
		_, err = testDB.Exec(ctx, "UPDATE urls SET is_deleted = true WHERE short_path = $1", "deleted123")
		require.NoError(t, err)

		// Пытаемся получить удаленную запись
		originalURL, err := repo.Get("deleted123")
		assert.Error(t, err)
		assert.Empty(t, originalURL)

		var goneErr *repository.ErrGone
		assert.ErrorAs(t, err, &goneErr)
		assert.Equal(t, "deleted123", goneErr.ShortPath)
	})
}

func TestPostgresRepository_Create(t *testing.T) {
	repo := setupDB(t)

	t.Run("successful create", func(t *testing.T) {
		entry := models.URLEntry{
			ShortPath: "test123",
			FullURL:   "https://test.com",
			UserID:    uuid.New().String(),
		}

		err := repo.Create(entry)
		require.NoError(t, err)

		// Проверяем, что запись создана
		originalURL, err := repo.Get("test123")
		require.NoError(t, err)
		assert.Equal(t, "https://test.com", originalURL)
	})

	t.Run("duplicate original_url", func(t *testing.T) {
		userID := uuid.New().String()
		entry1 := models.URLEntry{
			ShortPath: "first123",
			FullURL:   "https://duplicate.com",
			UserID:    userID,
		}
		err := repo.Create(entry1)
		require.NoError(t, err)

		// Пытаемся создать запись с тем же original_url
		entry2 := models.URLEntry{
			ShortPath: "second456",
			FullURL:   "https://duplicate.com",
			UserID:    userID,
		}
		err = repo.Create(entry2)
		assert.Error(t, err)

		var conflictErr *repository.ErrURLAlreadyExists
		assert.ErrorAs(t, err, &conflictErr)
		// Должен вернуться short_path первой записи
		assert.Equal(t, "first123", conflictErr.ShortPath)
	})
}

func TestPostgresRepository_CreateBatch(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	t.Run("successful batch create", func(t *testing.T) {
		userID := uuid.New().String()
		items := []models.URLEntry{
			{ShortPath: "batch1", FullURL: "https://batch1.com", UserID: userID},
			{ShortPath: "batch2", FullURL: "https://batch2.com", UserID: userID},
			{ShortPath: "batch3", FullURL: "https://batch3.com", UserID: userID},
		}

		err := repo.CreateBatch(ctx, items)
		require.NoError(t, err)

		// Проверяем, что все записи созданы
		for _, item := range items {
			originalURL, err := repo.Get(item.ShortPath)
			require.NoError(t, err)
			assert.Equal(t, item.FullURL, originalURL)
		}
	})

	t.Run("empty batch", func(t *testing.T) {
		err := repo.CreateBatch(ctx, []models.URLEntry{})
		require.NoError(t, err)
	})

	t.Run("batch with duplicates - ON CONFLICT DO NOTHING", func(t *testing.T) {
		userID := uuid.New().String()
		// Создаем первую запись
		entry1 := models.URLEntry{
			ShortPath: "existing",
			FullURL:   "https://existing.com",
			UserID:    userID,
		}
		err := repo.Create(entry1)
		require.NoError(t, err)

		// Пытаемся создать батч с дубликатом (тот же short_path)
		items := []models.URLEntry{
			{ShortPath: "existing", FullURL: "https://new.com", UserID: userID},
			{ShortPath: "new1", FullURL: "https://new1.com", UserID: userID},
		}

		err = repo.CreateBatch(ctx, items)
		// Должно работать без ошибки благодаря ON CONFLICT DO NOTHING
		require.NoError(t, err)

		// Проверяем, что старая запись не изменилась
		originalURL, err := repo.Get("existing")
		require.NoError(t, err)
		assert.Equal(t, "https://existing.com", originalURL)

		// Проверяем, что новая запись создана
		originalURL, err = repo.Get("new1")
		require.NoError(t, err)
		assert.Equal(t, "https://new1.com", originalURL)
	})
}

func TestPostgresRepository_GetByUserID(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	t.Run("successful get by user_id", func(t *testing.T) {
		userID1 := uuid.New().String()
		userID2 := uuid.New().String()

		// Создаем записи для первого пользователя
		items1 := []models.URLEntry{
			{ShortPath: "user1_1", FullURL: "https://user1_1.com", UserID: userID1},
			{ShortPath: "user1_2", FullURL: "https://user1_2.com", UserID: userID1},
		}
		err := repo.CreateBatch(ctx, items1)
		require.NoError(t, err)

		// Создаем записи для второго пользователя
		items2 := []models.URLEntry{
			{ShortPath: "user2_1", FullURL: "https://user2_1.com", UserID: userID2},
		}
		err = repo.CreateBatch(ctx, items2)
		require.NoError(t, err)

		// Получаем записи первого пользователя
		userURLs, err := repo.GetByUserID(ctx, userID1)
		require.NoError(t, err)
		assert.Len(t, userURLs, 2)

		// Проверяем содержимое
		shortPaths := make(map[string]string)
		for _, u := range userURLs {
			shortPaths[u.ShortPath] = u.OriginalURL
		}
		assert.Equal(t, "https://user1_1.com", shortPaths["user1_1"])
		assert.Equal(t, "https://user1_2.com", shortPaths["user1_2"])
		assert.NotContains(t, shortPaths, "user2_1") // Не должно быть записей другого пользователя
	})

	t.Run("empty result for nonexistent user", func(t *testing.T) {
		userURLs, err := repo.GetByUserID(ctx, uuid.New().String())
		require.NoError(t, err)
		assert.Empty(t, userURLs)
	})

	t.Run("filter by user_id - no foreign URLs", func(t *testing.T) {
		userID1 := uuid.New().String()
		userID2 := uuid.New().String()

		// Создаем записи для разных пользователей
		err := repo.Create(models.URLEntry{
			ShortPath: "foreign",
			FullURL:   "https://foreign.com",
			UserID:    userID2,
		})
		require.NoError(t, err)

		err = repo.Create(models.URLEntry{
			ShortPath: "own",
			FullURL:   "https://own.com",
			UserID:    userID1,
		})
		require.NoError(t, err)

		// Получаем записи первого пользователя
		userURLs, err := repo.GetByUserID(ctx, userID1)
		require.NoError(t, err)
		assert.Len(t, userURLs, 1)
		assert.Equal(t, "own", userURLs[0].ShortPath)
		assert.Equal(t, "https://own.com", userURLs[0].OriginalURL)
	})
}

func TestPostgresRepository_DeleteBatch(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		userID := uuid.New().String()

		// Создаем записи
		items := []models.URLEntry{
			{ShortPath: "delete1", FullURL: "https://delete1.com", UserID: userID},
			{ShortPath: "delete2", FullURL: "https://delete2.com", UserID: userID},
			{ShortPath: "keep", FullURL: "https://keep.com", UserID: userID},
		}
		err := repo.CreateBatch(ctx, items)
		require.NoError(t, err)

		// Удаляем первые две записи
		err = repo.DeleteBatch(ctx, userID, []string{"delete1", "delete2"})
		require.NoError(t, err)

		// Проверяем, что удаленные записи возвращают ErrGone
		_, err = repo.Get("delete1")
		var goneErr *repository.ErrGone
		assert.ErrorAs(t, err, &goneErr)

		_, err = repo.Get("delete2")
		assert.ErrorAs(t, err, &goneErr)

		// Проверяем, что не удаленная запись доступна
		originalURL, err := repo.Get("keep")
		require.NoError(t, err)
		assert.Equal(t, "https://keep.com", originalURL)
	})

	t.Run("delete only own URLs", func(t *testing.T) {
		userID1 := uuid.New().String()
		userID2 := uuid.New().String()

		// Создаем записи для разных пользователей
		err := repo.Create(models.URLEntry{
			ShortPath: "own",
			FullURL:   "https://own.com",
			UserID:    userID1,
		})
		require.NoError(t, err)

		err = repo.Create(models.URLEntry{
			ShortPath: "foreign",
			FullURL:   "https://foreign.com",
			UserID:    userID2,
		})
		require.NoError(t, err)

		// Пытаемся удалить чужую запись
		err = repo.DeleteBatch(ctx, userID1, []string{"foreign"})
		require.NoError(t, err) // Не должно быть ошибки

		// Проверяем, что чужая запись НЕ удалена
		originalURL, err := repo.Get("foreign")
		require.NoError(t, err)
		assert.Equal(t, "https://foreign.com", originalURL)

		// Проверяем, что своя запись доступна
		originalURL, err = repo.Get("own")
		require.NoError(t, err)
		assert.Equal(t, "https://own.com", originalURL)
	})

	t.Run("delete nonexistent URLs", func(t *testing.T) {
		userID := uuid.New().String()

		// Пытаемся удалить несуществующие записи
		err := repo.DeleteBatch(ctx, userID, []string{"nonexistent1", "nonexistent2"})
		require.NoError(t, err) // Не должно быть ошибки
	})

	t.Run("delete already deleted URLs", func(t *testing.T) {
		userID := uuid.New().String()

		// Создаем и удаляем запись
		err := repo.Create(models.URLEntry{
			ShortPath: "already_deleted",
			FullURL:   "https://deleted.com",
			UserID:    userID,
		})
		require.NoError(t, err)

		err = repo.DeleteBatch(ctx, userID, []string{"already_deleted"})
		require.NoError(t, err)

		// Пытаемся удалить еще раз
		err = repo.DeleteBatch(ctx, userID, []string{"already_deleted"})
		require.NoError(t, err) // Не должно быть ошибки
	})

	t.Run("empty delete list", func(t *testing.T) {
		userID := uuid.New().String()
		err := repo.DeleteBatch(ctx, userID, []string{})
		require.NoError(t, err)
	})
}

func TestPostgresRepository_Ping(t *testing.T) {
	repo := setupDB(t)
	ctx := context.Background()

	t.Run("successful ping", func(t *testing.T) {
		err := repo.Ping(ctx)
		require.NoError(t, err)
	})
}
