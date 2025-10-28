package grpc

import (
	"context"
	"encoding/json"
	"log"
	"net"

	pb "github.com/eyupaydin41/event-store/proto"
	"github.com/eyupaydin41/event-store/service"
	"google.golang.org/grpc"
)

// EventStoreServer - gRPC server implementation
// HTTP'deki Gin handler'ların karşılığı
type EventStoreServer struct {
	pb.UnimplementedEventStoreServiceServer // Forward compatibility için gerekli
	eventService                            *service.EventService
	snapshotService                         *service.SnapshotService
}

// NewEventStoreServer - Constructor
func NewEventStoreServer(eventService *service.EventService, snapshotService *service.SnapshotService) *EventStoreServer {
	return &EventStoreServer{
		eventService:    eventService,
		snapshotService: snapshotService,
	}
}

// GetAggregateEvents - RPC method implementation
// HTTP karşılığı: GET /events/aggregate/:id
func (s *EventStoreServer) GetAggregateEvents(
	ctx context.Context,
	req *pb.GetAggregateEventsRequest,
) (*pb.GetAggregateEventsResponse, error) {
	log.Printf("gRPC: GetAggregateEvents called for aggregate_id: %s", req.AggregateId)

	// HTTP'de: aggregateID := c.Param("id")
	aggregateID := req.AggregateId

	// Mevcut service'i kullan
	// fromVersion = 0 tüm event'leri getir
	events, err := s.eventService.GetEventsByAggregateID(aggregateID, 0)
	if err != nil {
		log.Printf("gRPC: Error fetching events: %v", err)
		// HTTP'de: c.JSON(500, ...)
		return nil, err
	}

	// Domain event'leri protobuf message'a dönüştür
	// HTTP'de: c.JSON(200, events)
	pbEvents := make([]*pb.Event, len(events))
	for i, event := range events {
		pbEvents[i] = &pb.Event{
			Id:          event.ID,
			EventType:   event.EventType,
			AggregateId: event.AggregateID,
			Version:     int32(event.Version),
			Timestamp:   event.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
			DataJson:    event.Payload,
		}
	}

	return &pb.GetAggregateEventsResponse{
		Events: pbEvents,
	}, nil
}

// GetAggregateWithSnapshot - Snapshot kullanarak aggregate state'ini getir
func (s *EventStoreServer) GetAggregateWithSnapshot(
	ctx context.Context,
	req *pb.GetAggregateWithSnapshotRequest,
) (*pb.GetAggregateWithSnapshotResponse, error) {
	log.Printf("gRPC: GetAggregateWithSnapshot called for aggregate_id: %s", req.AggregateId)

	aggregateID := req.AggregateId

	// Snapshot service ile aggregate'i yükle
	// Bu otomatik olarak:
	// 1. Snapshot varsa: snapshot + sonraki eventleri kullanır
	// 2. Snapshot yoksa: tüm eventleri kullanır
	aggregate, err := s.snapshotService.LoadAggregateWithSnapshot(aggregateID)
	if err != nil {
		log.Printf("gRPC: Error loading aggregate with snapshot: %v", err)
		return nil, err
	}

	// Aggregate'in state'ini JSON'a serialize et
	stateJSON, err := json.Marshal(aggregate)
	if err != nil {
		log.Printf("gRPC: Error marshaling aggregate state: %v", err)
		return nil, err
	}

	// Snapshot'tan mı yüklendi kontrol et
	hasSnapshot, _ := s.snapshotService.HasSnapshot(aggregateID)

	// EventCount: Toplam kaç event replay edildi
	eventsReplayed := aggregate.EventCount

	log.Printf("gRPC: Aggregate loaded - Version: %d, EventsReplayed: %d, FromSnapshot: %v",
		aggregate.Version, eventsReplayed, hasSnapshot)

	return &pb.GetAggregateWithSnapshotResponse{
		AggregateId:    aggregateID,
		Version:        aggregate.Version,
		StateJson:      string(stateJSON),
		FromSnapshot:   hasSnapshot,
		EventsReplayed: uint32(eventsReplayed),
	}, nil
}

// StartGRPCServer - gRPC server'ı başlat
// HTTP'de: router.Run(":8090")
func StartGRPCServer(port string, eventService *service.EventService, snapshotService *service.SnapshotService) error {
	// gRPC server oluştur
	grpcServer := grpc.NewServer()

	// Service'i register et
	pb.RegisterEventStoreServiceServer(grpcServer, NewEventStoreServer(eventService, snapshotService))

	// Listen
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	log.Printf("gRPC server starting on %s", port)
	return grpcServer.Serve(listener)
}
