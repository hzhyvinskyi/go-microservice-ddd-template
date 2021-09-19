//go:generate mockery --dir ../../domain --all --output ../../infrastructure/persistence/mocks

package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hzhyvinskyi/go-microservice-template/internal/app/application/pb"
	"github.com/hzhyvinskyi/go-microservice-template/internal/app/domain"
	"github.com/hzhyvinskyi/go-microservice-template/internal/app/infrastructure/persistence/mocks"
)

var testTemplate = &domain.Template{
	ID:        gofakeit.UUID(),
	Name:      gofakeit.Word(),
	CreatedAt: time.Now().String(),
}

func TestGet(t *testing.T) {
	templateRepositoryMock := &mocks.Repository{}
	templateServiceServer := &templateServiceServer{
		repository: templateRepositoryMock,
	}

	templateRepositoryMock.On("Get", mock.Anything, mock.AnythingOfType("string")).
		Return(testTemplate, nil)

	getTemplateResp, err := templateServiceServer.Get(context.Background(), &pb.GetTemplateReq{Id: testTemplate.ID})
	assert.NoError(t, err)

	want := &pb.Template{
		Id:        testTemplate.ID,
		Name:      testTemplate.Name,
		CreatedAt: testTemplate.CreatedAt,
	}
	got := getTemplateResp.Template

	assert.Equal(t, want, got)
}
