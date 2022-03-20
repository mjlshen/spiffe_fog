package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"

	"github.com/mjlshen/spiffe_fog/workload"
)

type spiffeWorkloadAPIServer struct {
	workload.UnimplementedSpiffeWorkloadAPIServer
}

// FetchJWTSVID fetches JWT-SVIDs for all SPIFFE identities the workload is entitled to
// for the requested audience(s). If an optional SPIFFE ID is requested, only
// the JWT-SVID for that SPIFFE ID is returned.
func (s *spiffeWorkloadAPIServer) FetchJWTSVID(ctx context.Context, req *workload.JWTSVIDRequest) (*workload.JWTSVIDResponse, error) {
	//TODO implement me
	if len(req.Audience) == 0 {
		return nil, status.Error(codes.InvalidArgument, "audience must be specified")
	}

	return &workload.JWTSVIDResponse{}, nil
}

// FetchJWTBundles fetches the JWT bundles, formatted as JWKS documents, keyed by the
// SPIFFE ID of the trust domain. As this information changes, subsequent
// messages will be streamed from the server.
func (s *spiffeWorkloadAPIServer) FetchJWTBundles(req *workload.JWTBundlesRequest, stream workload.SpiffeWorkloadAPI_FetchJWTBundlesServer) error {
	//TODO implement me
	return nil
}

// ValidateJWTSVID validates a JWT-SVID against the requested audience. Returns the
// SPIFFE ID of the JWT-SVID and JWT claims.
func (s *spiffeWorkloadAPIServer) ValidateJWTSVID(ctx context.Context, req *workload.ValidateJWTSVIDRequest) (*workload.ValidateJWTSVIDResponse, error) {
	//TODO implement me
	if req.Audience == "" {
		return nil, status.Error(codes.InvalidArgument, "audience must be specified")
	}
	if req.Svid == "" {
		return nil, status.Error(codes.InvalidArgument, "svid must be specified")
	}

	return &workload.ValidateJWTSVIDResponse{
		SpiffeId: "",
		Claims:   nil,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	workload.RegisterSpiffeWorkloadAPIServer(s, &spiffeWorkloadAPIServer{})
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
