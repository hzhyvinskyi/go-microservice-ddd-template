package dynamo

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/hzhyvinskyi/go-microservice-template/internal/app/domain"
)

const tableName = "templates"

type repository struct {
	dynamo *dynamodb.DynamoDB
}

func NewRepository(dbHandle *dynamodb.DynamoDB) *repository {
	return &repository{
		dynamo: dbHandle,
	}
}

func (r *repository) Get(ctx context.Context, id string) (*domain.Template, error) {
	logger := ctxzap.Extract(ctx)

	getItemOutput, err := r.dynamo.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		logger.Error("Failed to get template", zap.String("TemplateID", id), zap.Error(err))
		return nil, err
	}

	template := &domain.Template{}
	err = dynamodbattribute.UnmarshalMap(getItemOutput.Item, template)
	if err != nil {
		logger.Error(
			"Failed to unmarshal from a map of AttributeValues to the Template",
			zap.String("TemplateID", id),
			zap.Error(err),
		)
		return nil, err
	}

	return template, nil
}

func (r *repository) List(ctx context.Context, filter interface{}) ([]*domain.Template, error) {
	logger := ctxzap.Extract(ctx)

	transactGetItemsOutput, err := r.dynamo.TransactGetItemsWithContext(ctx, &dynamodb.TransactGetItemsInput{
		ReturnConsumedCapacity: nil,
		TransactItems:          nil,
	})
	if err != nil {
		logger.Error("Failed to get templates", zap.Error(err))
		return nil, err
	}

	templates := make([]*domain.Template, 0, len(transactGetItemsOutput.Responses))
	for _, response := range transactGetItemsOutput.Responses {
		template := &domain.Template{}
		err := dynamodbattribute.UnmarshalMap(response.Item, template)
		if err != nil {
			logger.Error(
				"Failed to unmarshal from a map of AttributeValues to the Template",
				zap.Error(err),
			)
			return nil, err
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (r *repository) Add(ctx context.Context, template *domain.Template) (*domain.Template, error) {
	_, err := r.dynamo.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(template.ID),
			},
			"Name": {
				S: aws.String(template.Name),
			},
			"CreatedAt": {
				S: aws.String(time.Now().Format("02-01-2006 15:04:05")),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		ctxzap.Error(ctx, "Failed to add new template", zap.String("TemplateID", template.ID), zap.Error(err))
		return nil, err
	}

	return template, nil
}

func (r *repository) Update(ctx context.Context, template *domain.Template) (*domain.Template, error) {
	_, err := r.dynamo.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("Name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":Name": {
				S: aws.String(template.Name),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(template.ID),
			},
		},
		TableName:        aws.String(tableName),
		UpdateExpression: aws.String("SET #N = :Name"),
	})
	if err != nil {
		ctxzap.Error(ctx, "Failed to update template", zap.String("TemplateID", template.ID), zap.Error(err))
		return nil, err
	}

	return template, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	_, err := r.dynamo.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		ctxzap.Error(ctx, "Failed to delete template", zap.String("TemplateID", id), zap.Error(err))
		return err
	}

	return nil
}
