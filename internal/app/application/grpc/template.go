package grpc

import (
	"context"

	"github.com/hzhyvinskyi/go-microservice-template/internal/app/application/pb"
	"github.com/hzhyvinskyi/go-microservice-template/internal/app/domain"
)

type templateServiceServer struct {
	repository domain.Repository
}

func NewTemplateServiceServer(repository domain.Repository) *templateServiceServer {
	return &templateServiceServer{
		repository: repository,
	}
}

func (s *templateServiceServer) Get(ctx context.Context, req *pb.GetTemplateReq) (*pb.GetTemplateResp, error) {
	template, err := s.repository.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	resp := &pb.GetTemplateResp{
		Template: &pb.Template{
			Id:        template.ID,
			Name:      template.Name,
			CreatedAt: template.CreatedAt,
		},
	}

	return resp, nil
}
