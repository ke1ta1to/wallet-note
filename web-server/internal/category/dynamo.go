package category

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	typeCategory = "Category"
	skPrefix     = "CATEGORY#"
)

func orgPK(orgID string) string           { return "ORG#" + orgID }
func categorySK(categoryID string) string { return skPrefix + categoryID }

type categoryItem struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	Type       string `dynamodbav:"type"`
	OrgID      string `dynamodbav:"org_id"`
	CategoryID string `dynamodbav:"category_id"`
	Name       string `dynamodbav:"name"`
	Kind       string `dynamodbav:"kind"`
	Color      string `dynamodbav:"color"`
	CreatedAt  string `dynamodbav:"created_at"`
}

func categoryItemFromModel(c *Category) categoryItem {
	return categoryItem{
		PK:         orgPK(c.OrgID),
		SK:         categorySK(c.ID),
		Type:       typeCategory,
		OrgID:      c.OrgID,
		CategoryID: c.ID,
		Name:       c.Name,
		Kind:       c.Kind,
		Color:      c.Color,
		CreatedAt:  c.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (i categoryItem) toModel() (*Category, error) {
	t, err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &Category{
		ID:        i.CategoryID,
		OrgID:     i.OrgID,
		Name:      i.Name,
		Kind:      i.Kind,
		Color:     i.Color,
		CreatedAt: t,
	}, nil
}

type CategoryRepo struct {
	db        *dynamodb.Client
	tableName string
}

func NewCategoryRepo(db *dynamodb.Client, tableName string) *CategoryRepo {
	return &CategoryRepo{db: db, tableName: tableName}
}

func (r *CategoryRepo) Put(ctx context.Context, c *Category) error {
	av, err := attributevalue.MarshalMap(categoryItemFromModel(c))
	if err != nil {
		return fmt.Errorf("marshal category: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      av,
	}); err != nil {
		return fmt.Errorf("put category: %w", err)
	}
	return nil
}

// ListByOrg returns categories for the org sorted ascending by SK
// (UUIDv7 → creation order, oldest first).
func (r *CategoryRepo) ListByOrg(ctx context.Context, orgID string) ([]*Category, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk": &ddbtypes.AttributeValueMemberS{Value: orgPK(orgID)},
			":sk": &ddbtypes.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	result := make([]*Category, 0, len(out.Items))
	for _, av := range out.Items {
		var item categoryItem
		if err := attributevalue.UnmarshalMap(av, &item); err != nil {
			return nil, fmt.Errorf("unmarshal category: %w", err)
		}
		c, err := item.toModel()
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}
