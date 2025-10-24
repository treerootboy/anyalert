package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/treerootboy/anyalert/pkg/notifier"
	pb "github.com/treerootboy/anyalert/proto"
)

// Server represents the gRPC server
type Server struct {
	pb.UnimplementedNotificationServiceServer
	manager *notifier.Manager
	addr    string
}

// NewServer creates a new gRPC server
func NewServer(manager *notifier.Manager, host string, port int) *Server {
	return &Server{
		manager: manager,
		addr:    fmt.Sprintf("%s:%d", host, port),
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationServiceServer(grpcServer, s)

	log.Printf("Starting gRPC server on %s", s.addr)
	return grpcServer.Serve(lis)
}

// Send sends a notification to a single channel
func (s *Server) Send(ctx context.Context, req *pb.SendRequest) (*pb.SendResponse, error) {
	msg := convertProtoMessage(req.Message)
	resp, err := s.manager.Send(ctx, req.Channel, msg)
	if err != nil {
		return &pb.SendResponse{
			Response: &pb.Response{
				Success: false,
				Error:   err.Error(),
			},
		}, nil
	}

	return &pb.SendResponse{
		Response: convertResponse(resp),
	}, nil
}

// Broadcast sends a notification to multiple channels
func (s *Server) Broadcast(ctx context.Context, req *pb.BroadcastRequest) (*pb.BroadcastResponse, error) {
	msg := convertProtoMessage(req.Message)
	results := s.manager.Broadcast(ctx, req.Channels, msg)

	protoResults := make(map[string]*pb.Response)
	for channel, resp := range results {
		protoResults[channel] = convertResponse(resp)
	}

	return &pb.BroadcastResponse{
		Results: protoResults,
	}, nil
}

// ListChannels returns available notification channels
func (s *Server) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	channels := s.manager.ListChannels()
	return &pb.ListChannelsResponse{
		Channels: channels,
	}, nil
}

// convertProtoMessage converts a proto Message to a notifier Message
func convertProtoMessage(msg *pb.Message) *notifier.Message {
	return &notifier.Message{
		To:       msg.To,
		Subject:  msg.Subject,
		Content:  msg.Content,
		Priority: msg.Priority,
		Metadata: msg.Metadata,
	}
}

// convertResponse converts a notifier Response to a proto Response
func convertResponse(resp *notifier.Response) *pb.Response {
	details := make(map[string]string)
	if resp.Details != nil {
		for k, v := range resp.Details {
			details[k] = fmt.Sprintf("%v", v)
		}
	}

	return &pb.Response{
		Success:   resp.Success,
		MessageId: resp.MessageID,
		Error:     resp.Error,
		Details:   details,
	}
}
