package service

import (
	"fmt"
	"log"
	"time"

	"github.com/eyupaydin41/event-store/model"
	"github.com/eyupaydin41/event-store/repository"
)

type EventService struct {
	repo *repository.EventRepository
}

func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) SaveEvent(event *model.Event) error {
	if event.AggregateID != "" {
		latestVersion, err := s.repo.GetLatestVersionForAggregate(event.AggregateID)
		if err != nil {
			log.Printf("error getting latest version for aggregate %s: %v", event.AggregateID, err)
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		event.Version = latestVersion + 1
	} else {
		event.Version = 1
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if err := s.repo.SaveEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	log.Printf("event saved: %s (type: %s, version: %d)", event.ID, event.EventType, event.Version)
	return nil
}

func (s *EventService) GetEvents(filter model.EventFilter) ([]*model.Event, error) {
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}

	log.Printf("retrieved %d events with filter", len(events))
	return events, nil
}

func (s *EventService) GetEventsByAggregateID(aggregateID string, fromVersion int) ([]*model.Event, error) {
	if aggregateID == "" {
		return nil, fmt.Errorf("aggregate_id is required")
	}

	filter := model.EventFilter{
		AggregateID: aggregateID,
		Limit:       1000,
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for aggregate %s: %w", aggregateID, err)
	}

	if fromVersion > 0 {
		var filtered []*model.Event
		for _, event := range events {
			if int(event.Version) >= fromVersion {
				filtered = append(filtered, event)
			}
		}
		events = filtered
	}

	log.Printf("retrieved %d events for aggregate: %s", len(events), aggregateID)
	return events, nil
}

func (s *EventService) GetEventsSince(since time.Time) ([]*model.Event, error) {
	if since.IsZero() {
		return nil, fmt.Errorf("since time is required")
	}

	filter := model.EventFilter{
		StartTime: since,
		Limit:     10000,
	}

	events, err := s.repo.GetEvents(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get events since %v: %w", since, err)
	}

	log.Printf("retrieved %d events since %v", len(events), since)
	return events, nil
}

func (s *EventService) CountEvents() (uint64, error) {
	count, err := s.repo.CountEvents()
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

func (s *EventService) GetLatestVersionForAggregate(aggregateID string) (uint32, error) {
	if aggregateID == "" {
		return 0, fmt.Errorf("aggregate_id is required")
	}

	version, err := s.repo.GetLatestVersionForAggregate(aggregateID)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest version for aggregate %s: %w", aggregateID, err)
	}

	return version, nil
}
