package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/repository"
	"github.com/google/uuid"
)

type SnapshotService struct {
	snapshotRepo *repository.SnapshotRepository
	eventRepo    *repository.EventRepository
}

func NewSnapshotService(snapshotRepo *repository.SnapshotRepository, eventRepo *repository.EventRepository) *SnapshotService {
	return &SnapshotService{
		snapshotRepo: snapshotRepo,
		eventRepo:    eventRepo,
	}
}

// CreateSnapshot - Aggregate için snapshot oluşturur
func (s *SnapshotService) CreateSnapshot(aggregateID string) error {
	// 1. Aggregate için tüm event'leri al
	filter := model.EventFilter{
		AggregateID: aggregateID,
		Limit:       10000,
	}

	events, err := s.eventRepo.GetEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to get events for snapshot: %w", err)
	}

	if len(events) == 0 {
		return fmt.Errorf("no events found for aggregate: %s", aggregateID)
	}

	// 2. Event'lerden aggregate state'ini oluştur
	aggregate := model.NewUserAggregate()
	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			return fmt.Errorf("failed to apply event: %w", err)
		}
	}

	// 3. Aggregate state'ini JSON'a çevir
	stateJSON, err := json.Marshal(aggregate)
	if err != nil {
		return fmt.Errorf("failed to marshal aggregate state: %w", err)
	}

	// 4. Snapshot'ı kaydet
	snapshot := &model.Snapshot{
		ID:          uuid.New().String(),
		AggregateID: aggregateID,
		Version:     aggregate.Version,
		State:       string(stateJSON),
		CreatedAt:   time.Now(),
	}

	if err := s.snapshotRepo.SaveSnapshot(snapshot); err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	log.Printf("Snapshot created for aggregate %s at version %d", aggregateID, aggregate.Version)
	return nil
}

// LoadAggregateWithSnapshot - Snapshot kullanarak aggregate'i yükler
// Önce en son snapshot'ı alır, sonra snapshot'tan sonraki event'leri uygular
func (s *SnapshotService) LoadAggregateWithSnapshot(aggregateID string) (*model.UserAggregate, error) {
	// 1. En son snapshot'ı al
	snapshot, err := s.snapshotRepo.GetLatestSnapshot(aggregateID)

	var aggregate *model.UserAggregate

	if err != nil {
		// Snapshot yok, tüm event'leri yükle
		log.Printf("No snapshot found for aggregate %s, loading from all events", aggregateID)
		return s.loadFromAllEvents(aggregateID)
	}

	// 2. Snapshot'tan aggregate'i deserialize et
	aggregate = model.NewUserAggregate()
	if err := json.Unmarshal([]byte(snapshot.State), aggregate); err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot state: %w", err)
	}

	log.Printf("Loaded snapshot for aggregate %s at version %d", aggregateID, snapshot.Version)

	// 3. Snapshot'tan sonraki event'leri al
	events, err := s.eventRepo.GetEventsAfterVersion(aggregateID, snapshot.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get events after snapshot: %w", err)
	}

	// 4. Snapshot'tan sonraki event'leri uygula
	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
	}

	log.Printf("Applied %d events after snapshot for aggregate %s", len(events), aggregateID)
	return aggregate, nil
}

// loadFromAllEvents - Tüm event'lerden aggregate'i yükler (snapshot yoksa)
func (s *SnapshotService) loadFromAllEvents(aggregateID string) (*model.UserAggregate, error) {
	filter := model.EventFilter{
		AggregateID: aggregateID,
		Limit:       10000,
	}

	events, err := s.eventRepo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	if len(events) == 0 {
		return nil, fmt.Errorf("aggregate not found: %s", aggregateID)
	}

	aggregate := model.NewUserAggregate()
	for _, event := range events {
		if err := aggregate.ApplyEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
	}

	return aggregate, nil
}

// LoadAggregateAtVersion - Belirli bir version'daki aggregate state'ini yükler
func (s *SnapshotService) LoadAggregateAtVersion(aggregateID string, targetVersion uint32) (*model.UserAggregate, error) {
	// 1. Target version'dan önce veya eşit olan en son snapshot'ı al
	snapshot, err := s.snapshotRepo.GetSnapshotAtVersion(aggregateID, targetVersion)

	var aggregate *model.UserAggregate
	var fromVersion uint32 = 0

	if err != nil {
		// Snapshot yok, baştan başla
		aggregate = model.NewUserAggregate()
	} else {
		// Snapshot'tan başla
		aggregate = model.NewUserAggregate()
		if err := json.Unmarshal([]byte(snapshot.State), aggregate); err != nil {
			return nil, fmt.Errorf("failed to unmarshal snapshot state: %w", err)
		}
		fromVersion = snapshot.Version
	}

	// 2. Snapshot'tan target version'a kadar olan event'leri al
	filter := model.EventFilter{
		AggregateID: aggregateID,
		Limit:       10000,
	}

	events, err := s.eventRepo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	// 3. Target version'a kadar event'leri uygula
	for _, event := range events {
		if event.Version > fromVersion && event.Version <= targetVersion {
			if err := aggregate.ApplyEvent(event); err != nil {
				return nil, fmt.Errorf("failed to apply event: %w", err)
			}
		}
	}

	return aggregate, nil
}

// ShouldCreateSnapshot - Snapshot oluşturulmalı mı kontrol eder
// Her N event'te bir snapshot oluştur (örneğin her 50 event)
func (s *SnapshotService) ShouldCreateSnapshot(aggregateID string, snapshotInterval uint32) (bool, error) {
	latestVersion, err := s.eventRepo.GetLatestVersionForAggregate(aggregateID)
	if err != nil {
		return false, err
	}

	// Snapshot var mı kontrol et
	hasSnapshot, err := s.snapshotRepo.HasSnapshot(aggregateID)
	if err != nil {
		return false, err
	}

	if !hasSnapshot {
		// İlk snapshot - 10 event'ten sonra oluştur
		return latestVersion >= 10, nil
	}

	// Son snapshot'ı al
	snapshot, err := s.snapshotRepo.GetLatestSnapshot(aggregateID)
	if err != nil {
		return false, err
	}

	// Snapshot'tan sonra yeterli event var mı?
	eventsSinceSnapshot := latestVersion - snapshot.Version
	return eventsSinceSnapshot >= snapshotInterval, nil
}

// AutoCreateSnapshots - Belirli bir aggregate için otomatik snapshot oluşturur
func (s *SnapshotService) AutoCreateSnapshots(aggregateID string, snapshotInterval uint32) error {
	shouldCreate, err := s.ShouldCreateSnapshot(aggregateID, snapshotInterval)
	if err != nil {
		return err
	}

	if shouldCreate {
		return s.CreateSnapshot(aggregateID)
	}

	return nil
}
