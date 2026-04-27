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
	"github.com/ke1ta1to/wallet-note/internal/platform/idgen"
)

type Service struct {
	db        *dynamodb.Client
	tableName string
}

func NewService(db *dynamodb.Client, tableName string) *Service {
	return &Service{db: db, tableName: tableName}
}

// CreateOrgWithMembership writes Organization meta and the owner Membership
// in a single TransactWriteItems so neither can exist without the other.
func (s *Service) CreateOrgWithMembership(ctx context.Context, ownerUserID, name string) (*Organization, error) {
	now := time.Now().UTC()
	org := &Organization{
		ID:        idgen.NewID(),
		Name:      name,
		CreatedAt: now,
		CreatedBy: ownerUserID,
	}
	mem := &auth.Membership{
		UserID:   ownerUserID,
		OrgID:    org.ID,
		Role:     auth.RoleOwner,
		JoinedAt: now,
	}

	orgAV, err := attributevalue.MarshalMap(orgItemFromModel(org))
	if err != nil {
		return nil, fmt.Errorf("marshal org: %w", err)
	}
	memAV, err := attributevalue.MarshalMap(membershipItemFromModel(mem))
	if err != nil {
		return nil, fmt.Errorf("marshal membership: %w", err)
	}

	_, err = s.db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{
				Put: &ddbtypes.Put{
					TableName:           &s.tableName,
					Item:                orgAV,
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
				},
			},
			{
				Put: &ddbtypes.Put{
					TableName:           &s.tableName,
					Item:                memAV,
					ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("transact write: %w", err)
	}
	return org, nil
}
