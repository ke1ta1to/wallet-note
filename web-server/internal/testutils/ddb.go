// Package testutils provides helpers for request tests.
package testutils

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const TableName = "wallet-note-test"

// DDB connects to DynamoDB Local and ensures the test table exists.
// Requires `docker compose up dynamodb-local`.
func DDB(t *testing.T) *dynamodb.Client {
	t.Helper()
	ctx := context.Background()

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("ap-northeast-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	db := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	if _, err := db.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(TableName),
	}); err != nil {
		if _, ok := errors.AsType[*ddbtypes.ResourceNotFoundException](err); !ok {
			t.Fatalf("describe test table: %v", err)
		}
		if err := createTable(ctx, db); err != nil {
			t.Fatalf("create test table: %v", err)
		}
	}
	return db
}

// createTable mirrors infrastructure/modules/dynamodb/main.tf. Keep schema
// (PK, SK, GSI1PK, GSI1SK, GSI1) in sync when the production table changes.
// TTL / PITR are intentionally omitted: they are not exercised by tests.
func createTable(ctx context.Context, db *dynamodb.Client) error {
	_, err := db.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(TableName),
		AttributeDefinitions: []ddbtypes.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: ddbtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: ddbtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: ddbtypes.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: ddbtypes.ScalarAttributeTypeS},
		},
		KeySchema: []ddbtypes.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: ddbtypes.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: ddbtypes.KeyTypeRange},
		},
		GlobalSecondaryIndexes: []ddbtypes.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []ddbtypes.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: ddbtypes.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: ddbtypes.KeyTypeRange},
				},
				Projection: &ddbtypes.Projection{
					ProjectionType: ddbtypes.ProjectionTypeAll,
				},
			},
		},
		BillingMode: ddbtypes.BillingModePayPerRequest,
	})
	return err
}

func ResetTable(t *testing.T, db *dynamodb.Client) {
	t.Helper()
	ctx := context.Background()

	var lastKey map[string]ddbtypes.AttributeValue
	for {
		out, err := db.Scan(ctx, &dynamodb.ScanInput{
			TableName:            aws.String(TableName),
			ProjectionExpression: aws.String("PK, SK"),
			ExclusiveStartKey:    lastKey,
		})
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		for _, item := range out.Items {
			if _, err := db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(TableName),
				Key:       item,
			}); err != nil {
				t.Fatalf("delete item: %v", err)
			}
		}
		if out.LastEvaluatedKey == nil {
			break
		}
		lastKey = out.LastEvaluatedKey
	}
}
