package grpc

import (
	"context"
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
}

// NewEventStoreServer - Constructor
func NewEventStoreServer(eventService *service.EventService) *EventStoreServer {
	return &EventStoreServer{
		eventService: eventService,
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

// StartGRPCServer - gRPC server'ı başlat
// HTTP'de: router.Run(":8090")
func StartGRPCServer(port string, eventService *service.EventService) error {
	// gRPC server oluştur
	grpcServer := grpc.NewServer()

	// Service'i register et
	pb.RegisterEventStoreServiceServer(grpcServer, NewEventStoreServer(eventService))

	// Listen
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	log.Printf("gRPC server starting on %s", port)
	return grpcServer.Serve(listener)
}
