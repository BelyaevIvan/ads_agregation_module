package service

import (
	"sort"
	"strconv"
	"sync"
	"time"

	"brandhunt/api-service/internal/model"
	"brandhunt/api-service/internal/repository"
)

// cacheTTL — как долго данные считаются свежими. Список размеров меняется
// редко, поэтому 5 минут — разумный компромисс между актуальностью и нагрузкой.
const cacheTTL = 5 * time.Minute

type FiltersService struct {
	repo *repository.FiltersRepo

	mu        sync.RWMutex
	cached    *model.SizeFilters
	expiresAt time.Time
}

func NewFiltersService(repo *repository.FiltersRepo) *FiltersService {
	return &FiltersService{repo: repo}
}

// GetSizes возвращает уникальные размеры из активных объявлений.
// Результат кэшируется на 5 минут, чтобы не долбить БД одним и тем же запросом.
func (s *FiltersService) GetSizes() (*model.SizeFilters, error) {
	s.mu.RLock()
	if s.cached != nil && time.Now().Before(s.expiresAt) {
		out := *s.cached
		s.mu.RUnlock()
		return &out, nil
	}
	s.mu.RUnlock()

	rus, eu, us, err := s.repo.GetDistinctSizes()
	if err != nil {
		return nil, err
	}

	sortSizes(rus)
	sortSizes(eu)
	sortSizes(us)

	result := &model.SizeFilters{SizeRus: rus, SizeEU: eu, SizeUS: us}

	s.mu.Lock()
	s.cached = result
	s.expiresAt = time.Now().Add(cacheTTL)
	s.mu.Unlock()

	copy := *result
	return &copy, nil
}

// sortSizes сортирует строковые размеры численно (42 < 43 < 44, 8 < 8.5 < 9).
// Нечисловые значения (если вдруг попали из LLM) отправляются в конец
// и сортируются лексикографически между собой.
func sortSizes(sizes []string) {
	sort.SliceStable(sizes, func(i, j int) bool {
		a, errA := strconv.ParseFloat(sizes[i], 64)
		b, errB := strconv.ParseFloat(sizes[j], 64)
		switch {
		case errA == nil && errB == nil:
			return a < b
		case errA == nil:
			return true
		case errB == nil:
			return false
		default:
			return sizes[i] < sizes[j]
		}
	})
}
