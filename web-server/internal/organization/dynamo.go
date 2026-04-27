package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ke1ta1to/wallet-note/internal/auth"
	"github.com/ke1ta1to/wallet-note/internal/shared/apperror"
)

const (
	typeOrganization = "Organization"
	typeMembership   = "Membership"
	skMeta           = "META"
)

func orgPK(orgID string) string   { return "ORG#" + orgID }
func userPK(userID string) string { return "USER#" + userID }
func orgSK(orgID string) string   { return "ORG#" + orgID }

type orgItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	Type      string `dynamodbav:"type"`
	OrgID     string `dynamodbav:"orgId"`
	Name      string `dynamodbav:"name"`
	CreatedAt string `dynamodbav:"createdAt"`
	CreatedBy string `dynamodbav:"createdBy"`
}

func orgItemFromModel(o *Organization) orgItem {
	return orgItem{
		PK:        orgPK(o.OrgID),
		SK:        skMeta,
		Type:      typeOrganization,
		OrgID:     o.OrgID,
		Name:      o.Name,
		CreatedAt: o.CreatedAt.UTC().Format(time.RFC3339),
		CreatedBy: o.CreatedBy,
	}
}

func (i orgItem) toModel() (*Organization, error) {
	t, err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse createdAt: %w", err)
	}
	return &Organization{
		OrgID:     i.OrgID,
		Name:      i.Name,
		CreatedAt: t,
		CreatedBy: i.CreatedBy,
	}, nil
}

type membershipItem struct {
	PK       string `dynamodbav:"PK"`
	SK       string `dynamodbav:"SK"`
	GSI1PK   string `dynamodbav:"GSI1PK"`
	GSI1SK   string `dynamodbav:"GSI1SK"`
	Type     string `dynamodbav:"type"`
	UserID   string `dynamodbav:"userId"`
	OrgID    string `dynamodbav:"orgId"`
	Role     string `dynamodbav:"role"`
	JoinedAt string `dynamodbav:"joinedAt"`
}

func membershipItemFromModel(m *auth.Membership) membershipItem {
	return membershipItem{
		PK:       userPK(m.UserID),
		SK:       orgSK(m.OrgID),
		GSI1PK:   orgPK(m.OrgID),
		GSI1SK:   userPK(m.UserID),
		Type:     typeMembership,
		UserID:   m.UserID,
		OrgID:    m.OrgID,
		Role:     m.Role,
		JoinedAt: m.JoinedAt.UTC().Format(time.RFC3339),
	}
}

func (i membershipItem) toModel() (*auth.Membership, error) {
	t, err := time.Parse(time.RFC3339, i.JoinedAt)
	if err != nil {
		return nil, fmt.Errorf("parse joinedAt: %w", err)
	}
	return &auth.Membership{
		UserID:   i.UserID,
		OrgID:    i.OrgID,
		Role:     i.Role,
		JoinedAt: t,
	}, nil
}

type DynamoOrgRepository struct {
	db        *dynamodb.Client
	tableName string
}

func NewDynamoOrgRepository(db *dynamodb.Client, tableName string) *DynamoOrgRepository {
	return &DynamoOrgRepository{db: db, tableName: tableName}
}

func (r *DynamoOrgRepository) GetOrg(ctx context.Context, orgID string) (*Organization, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: orgPK(orgID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: skMeta},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get org: %w", err)
	}
	if out.Item == nil {
		return nil, apperror.ErrNotFound
	}
	var item orgItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal org: %w", err)
	}
	return item.toModel()
}

// BatchGetOrgs returns orgs in input order; missing items are silently skipped.
func (r *DynamoOrgRepository) BatchGetOrgs(ctx context.Context, orgIDs []string) ([]*Organization, error) {
	if len(orgIDs) == 0 {
		return nil, nil
	}
	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(orgIDs))
	for _, id := range orgIDs {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: orgPK(id)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: skMeta},
		})
	}
	out, err := r.db.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			r.tableName: {Keys: keys},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("batch get orgs: %w", err)
	}
	raw := out.Responses[r.tableName]
	byID := make(map[string]*Organization, len(raw))
	for _, av := range raw {
		var item orgItem
		if err := attributevalue.UnmarshalMap(av, &item); err != nil {
			return nil, fmt.Errorf("unmarshal org: %w", err)
		}
		o, err := item.toModel()
		if err != nil {
			return nil, err
		}
		byID[o.OrgID] = o
	}
	result := make([]*Organization, 0, len(orgIDs))
	for _, id := range orgIDs {
		if o, ok := byID[id]; ok {
			result = append(result, o)
		}
	}
	return result, nil
}

type DynamoMembershipRepository struct {
	db        *dynamodb.Client
	tableName string
}

func NewDynamoMembershipRepository(db *dynamodb.Client, tableName string) *DynamoMembershipRepository {
	return &DynamoMembershipRepository{db: db, tableName: tableName}
}

func (r *DynamoMembershipRepository) GetMembership(ctx context.Context, userID, orgID string) (*auth.Membership, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.tableName,
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: userPK(userID)},
			"SK": &ddbtypes.AttributeValueMemberS{Value: orgSK(orgID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get membership: %w", err)
	}
	if out.Item == nil {
		return nil, apperror.ErrNotFound
	}
	var item membershipItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return nil, fmt.Errorf("unmarshal membership: %w", err)
	}
	return item.toModel()
}

func (r *DynamoMembershipRepository) ListByUser(ctx context.Context, userID string) ([]*auth.Membership, error) {
	out, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              &r.tableName,
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":pk": &ddbtypes.AttributeValueMemberS{Value: userPK(userID)},
			":sk": &ddbtypes.AttributeValueMemberS{Value: "ORG#"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("query memberships: %w", err)
	}
	result := make([]*auth.Membership, 0, len(out.Items))
	for _, av := range out.Items {
		var item membershipItem
		if err := attributevalue.UnmarshalMap(av, &item); err != nil {
			return nil, fmt.Errorf("unmarshal membership: %w", err)
		}
		m, err := item.toModel()
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}
