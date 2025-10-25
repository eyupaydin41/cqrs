package service

import (
	"fmt"
	"log"
	"time"

	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/repository"
)

// ReplayService - Event replay ve time travel işlemleri
type ReplayService struct {
	repo *repository.EventRepository
}

func NewReplayService(repo *repository.EventRepository) *ReplayService {
	return &ReplayService{repo: repo}
}

// ReplayUserState - Belirli bir user'ın mevcut durumunu event'lerden reconstruct eder
func (s *ReplayService) ReplayUserState(userID string) (*model.UserAggregate, error) {
	// Kullanıcının tüm event'lerini al
	filter := model.EventFilter{
		AggregateID: userID,
		Limit:       10000, // Tüm event'leri al
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for user: %s", userID)
	}

	// Boş aggregate ile başla
	aggregate := model.NewUserAggregate()

	// Tüm event'leri sırayla uygula
	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			log.Printf("Warning: failed to apply event %s: %v", event.ID, err)
			continue
		}
	}

	log.Printf("Replayed %d events for user %s (version: %d)", aggregate.EventCount, userID, aggregate.Version)
	return aggregate, nil
}

// ReplayUserStateAt - Belirli bir zamandaki user state'ini gösterir (TIME TRAVEL!)
func (s *ReplayService) ReplayUserStateAt(userID string, pointInTime time.Time) (*model.UserAggregate, error) {
	// Belirli zamana kadar olan event'leri al
	filter := model.EventFilter{
		AggregateID: userID,
		EndTime:     pointInTime,
		Limit:       10000,
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for user %s before %v", userID, pointInTime)
	}

	// Boş aggregate ile başla
	aggregate := model.NewUserAggregate()

	// Zamana kadar olan event'leri uygula
	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			log.Printf("Warning: failed to apply event %s: %v", event.ID, err)
			continue
		}
	}

	log.Printf("Time travel: Replayed %d events for user %s at %v (version: %d)",
		aggregate.EventCount, userID, pointInTime, aggregate.Version)
	return aggregate, nil
}

// GetUserHistory - Kullanıcının tüm değişiklik geçmişini döner
func (s *ReplayService) GetUserHistory(userID string) ([]*model.UserAggregate, error) {
	filter := model.EventFilter{
		AggregateID: userID,
		Limit:       10000,
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("no events found for user: %s", userID)
	}

	// Her event sonrası state'i kaydet
	var history []*model.UserAggregate
	aggregate := model.NewUserAggregate()

	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			log.Printf("Warning: failed to apply event %s: %v", event.ID, err)
			continue
		}

		// State'in bir kopyasını kaydet
		snapshot := *aggregate
		history = append(history, &snapshot)
	}

	log.Printf("Retrieved %d state snapshots for user %s", len(history), userID)
	return history, nil
}

// CompareStates - İki farklı zamandaki state'leri karşılaştırır
func (s *ReplayService) CompareStates(userID string, time1, time2 time.Time) (before, after *model.UserAggregate, err error) {
	before, err = s.ReplayUserStateAt(userID, time1)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get state at %v: %w", time1, err)
	}

	after, err = s.ReplayUserStateAt(userID, time2)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get state at %v: %w", time2, err)
	}

	return before, after, nil
}
