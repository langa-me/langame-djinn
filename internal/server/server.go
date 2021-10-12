package server

import (
	"context"
	"io"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	api "github.com/langa-me/langame-djinn/internal/djinn"
	"github.com/langa-me/langame-djinn/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ api.ConversationMagnifierServer = (*server)(nil)
var App *firebase.App

type server struct {
	api.UnimplementedConversationMagnifierServer
}

func NewServer() *server {
	return &server{
	}
}

// RouteChat receives a stream of message/location pairs, and responds with a stream of all
// previous messages at each of those locations.
func (s *server) Magnify(stream api.ConversationMagnifier_MagnifyServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("Received message: %v", in.Text)
		magnifiedText, err := service.Query(in.Text)
		if err != nil {
			return err
		}
		log.Printf("Magnified message: %v", magnifiedText)
		if err := stream.Send(&api.MagnificationResponse{
			Text: in.Text,
			Type: &api.MagnificationResponse_SentimentResponse{
				SentimentResponse: magnifiedText,
			},
		}); err != nil {
			return err
		}
	}
}


func parseToken(ctx context.Context, token string) (*auth.Token, error) {
	client, err := App.Auth(ctx)
	if err != nil {
		return nil, err
	}

	nt, err := client.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}

	log.Printf("Verified ID token: %v\n", token)
	return nt, nil
}

type AuthToken string
func AuthFunc(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	tokenInfo, err := parseToken(ctx, token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
	}
	grpc_ctxtags.Extract(ctx).Set("auth.sub", tokenInfo.UID)

	newCtx := context.WithValue(ctx, AuthToken("tokenInfo"), tokenInfo)

	return newCtx, nil
}
