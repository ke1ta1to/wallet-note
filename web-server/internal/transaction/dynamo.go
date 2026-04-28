package transaction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ke1ta1to/wallet-note/internal/platform/apperror"
)

const (
	typeTransaction = "Transaction"
	skPrefix        = "TX#"
)

func orgPK(orgID string) string                { return "ORG#" + orgID }
func txSK(date, txID string) string            { return skPrefix + date + "#" + txID }
func txGSI1PK(orgID, categoryID string) string { return "ORG#" + orgID + "#CAT#" + categoryID }
func txGSI1SK(date string) string              { return skPrefix + date }

type transactionItem struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	GSI1PK     string `dynamodbav:"GSI1PK"`
	GSI1SK     string `dynamodbav:"GSI1SK"`
	Type       string `dynamodbav:"type"`
	OrgID      string `dynamodbav:"org_id"`
	TxID       string `dynamodbav:"tx_id"`
	CategoryID string `dynamodbav:"category_id"`
	Amount     int64  `dynamodbav:"amount"`
	Date       string `dynamodbav:"date"`
	Memo       string `dynamodbav:"memo"`
	CreatedAt  string `dynamodbav:"created_at"`
	CreatedBy  string `dynamodbav:"created_by"`
}

func transactionItemFromModel(t *Transaction) transactionItem {
	return transactionItem{
		PK:         orgPK(t.OrgID),
		SK:         txSK(t.Date, t.ID),
		GSI1PK:     txGSI1PK(t.OrgID, t.CategoryID),
		GSI1SK:     txGSI1SK(t.Date),
		Type:       typeTransaction,
		OrgID:      t.OrgID,
		TxID:       t.ID,
		CategoryID: t.CategoryID,
		Amount:     t.Amount,
		Date:       t.Date,
		Memo:       t.Memo,
		CreatedAt:  t.CreatedAt.UTC().Format(time.RFC3339),
		CreatedBy:  t.CreatedBy,
	}
}

func (i transactionItem) toModel() (*Transaction, error) {
	t, err := time.Parse(time.RFC3339, i.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &Transaction{
		ID:         i.TxID,
		OrgID:      i.OrgID,
		CategoryID: i.CategoryID,
		Amount:     i.Amount,
		Date:       i.Date,
		Memo:       i.Memo,
		CreatedAt:  t,
		CreatedBy:  i.CreatedBy,
	}, nil
}

type TxRepo struct {
	db        *dynamodb.Client
	tableName string
}

func NewTxRepo(db *dynamodb.Client, tableName string) *TxRepo {
	return &TxRepo{db: db, tableName: tableName}
}

func (r *TxRepo) Put(ctx context.Context, t *Transaction) error {
	av, err := attributevalue.MarshalMap(transactionItemFromModel(t))
	if err != nil {
		return fmt.Errorf("marshal tx: %w", err)
	}
	if _, err := r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.tableName,
		Item:      av,
	}); err != nil {
		return fmt.Errorf("put tx: %w", err)
	}
	return nil
}

type ListByMonthInput struct {
	OrgID      string
	Month      string // YYYY-MM
	CategoryID string // optional, scopes via GSI1 when set
	Cursor     string // optional, base64 LastEvaluatedKey
}

type ListByMonthOutput struct {
	Items      []*Transaction
	NextCursor string
}

func (r *TxRepo) ListByMonth(ctx context.Context, in ListByMonthInput) (*ListByMonthOutput, error) {
	skPattern := skPrefix + in.Month
	var input *dynamodb.QueryInput
	if in.CategoryID != "" {
		input = &dynamodb.QueryInput{
			TableName:              &r.tableName,
			IndexName:              aws.String("GSI1"),
			KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":pk": &ddbtypes.AttributeValueMemberS{Value: txGSI1PK(in.OrgID, in.CategoryID)},
				":sk": &ddbtypes.AttributeValueMemberS{Value: skPattern},
			},
		}
	} else {
		input = &dynamodb.QueryInput{
			TableName:              &r.tableName,
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
			ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
				":pk": &ddbtypes.AttributeValueMemberS{Value: orgPK(in.OrgID)},
				":sk": &ddbtypes.AttributeValueMemberS{Value: skPattern},
			},
		}
	}
	if in.Cursor != "" {
		key, err := decodeCursor(in.Cursor)
		if err != nil {
			return nil, err
		}
		input.ExclusiveStartKey = key
	}

	out, err := r.db.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("query txs: %w", err)
	}
	items := make([]*Transaction, 0, len(out.Items))
	for _, av := range out.Items {
		var item transactionItem
		if err := attributevalue.UnmarshalMap(av, &item); err != nil {
			return nil, fmt.Errorf("unmarshal tx: %w", err)
		}
		m, err := item.toModel()
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	nextCursor := ""
	if len(out.LastEvaluatedKey) > 0 {
		nextCursor, err = encodeCursor(out.LastEvaluatedKey)
		if err != nil {
			return nil, err
		}
	}
	return &ListByMonthOutput{Items: items, NextCursor: nextCursor}, nil
}

// Cursor is a base64-encoded JSON map of string→string. All key attributes in
// our schema are strings, so this round-trips DDB's LastEvaluatedKey safely.
func encodeCursor(key map[string]ddbtypes.AttributeValue) (string, error) {
	plain := make(map[string]string, len(key))
	for k, v := range key {
		s, ok := v.(*ddbtypes.AttributeValueMemberS)
		if !ok {
			return "", fmt.Errorf("cursor: unsupported type for %s", k)
		}
		plain[k] = s.Value
	}
	b, err := json.Marshal(plain)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeCursor(cursor string) (map[string]ddbtypes.AttributeValue, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, apperror.ErrInvalidInput
	}
	var plain map[string]string
	if err := json.Unmarshal(b, &plain); err != nil {
		return nil, apperror.ErrInvalidInput
	}
	out := make(map[string]ddbtypes.AttributeValue, len(plain))
	for k, v := range plain {
		out[k] = &ddbtypes.AttributeValueMemberS{Value: v}
	}
	return out, nil
}
