package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"go-serverless-api-terraform/internal/models"
)

// Repository defines CRUD operations for orders and order items.
type Repository interface {
	// Orders
	CreateOrder(ctx context.Context, o *models.Order) error
	GetOrder(ctx context.Context, id string) (*models.Order, error)
	ListOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrder(ctx context.Context, o *models.Order) error
	DeleteOrder(ctx context.Context, id string) error

	// Order items
	CreateOrderItem(ctx context.Context, it *models.OrderItem) error
	GetOrderItem(ctx context.Context, orderID, id string) (*models.OrderItem, error)
	ListOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error)
	UpdateOrderItem(ctx context.Context, it *models.OrderItem) error
	DeleteOrderItem(ctx context.Context, orderID, id string) error
}

// DynamoRepository implements Repository using AWS DynamoDB.
type DynamoRepository struct {
	db              *dynamodb.Client
	ordersTable     string
	orderItemsTable string
}

func NewDynamoRepository(db *dynamodb.Client, ordersTable, orderItemsTable string) *DynamoRepository {
	return &DynamoRepository{db: db, ordersTable: ordersTable, orderItemsTable: orderItemsTable}
}

// Orders
func (r *DynamoRepository) CreateOrder(ctx context.Context, o *models.Order) error {
	if o == nil {
		return errors.New("order is nil")
	}
	item, err := attributevalue.MarshalMap(o)
	if err != nil {
		return err
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.ordersTable,
		Item:                item,
		ConditionExpression: awsString("attribute_not_exists(id)"),
	})
	return err
}

func (r *DynamoRepository) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	res, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.ordersTable,
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
	})
	if err != nil {
		return nil, err
	}
	if res.Item == nil {
		return nil, nil
	}
	var o models.Order
	if err := attributevalue.UnmarshalMap(res.Item, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *DynamoRepository) ListOrders(ctx context.Context) ([]models.Order, error) {
	res, err := r.db.Scan(ctx, &dynamodb.ScanInput{TableName: &r.ordersTable})
	if err != nil {
		return nil, err
	}
	var out []models.Order
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *DynamoRepository) UpdateOrder(ctx context.Context, o *models.Order) error {
	if o == nil {
		return errors.New("order is nil")
	}
	item, err := attributevalue.MarshalMap(o)
	if err != nil {
		return err
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{TableName: &r.ordersTable, Item: item})
	return err
}

func (r *DynamoRepository) DeleteOrder(ctx context.Context, id string) error {
	// delete order items first in parallel with bounded concurrency
	items, err := r.ListOrderItems(ctx, id)
	if err != nil {
		return err
	}
	sem := make(chan struct{}, 8) // concurrency limit
	var wg sync.WaitGroup
	var firstErr error
	var mu sync.Mutex
	for _, it := range items {
		wg.Add(1)
		it := it
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := r.DeleteOrderItem(ctx, id, it.ID); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	_, err = r.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.ordersTable,
		Key:       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: id}},
	})
	return err
}

// Order Items (PK: order_id, SK: id)
func (r *DynamoRepository) CreateOrderItem(ctx context.Context, it *models.OrderItem) error {
	if it == nil {
		return errors.New("order item is nil")
	}
	item, err := attributevalue.MarshalMap(it)
	if err != nil {
		return err
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.orderItemsTable,
		Item:                item,
		ConditionExpression: awsString("attribute_not_exists(order_id) AND attribute_not_exists(id)"),
	})
	return err
}

func (r *DynamoRepository) GetOrderItem(ctx context.Context, orderID, id string) (*models.OrderItem, error) {
	res, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.orderItemsTable,
		Key: map[string]types.AttributeValue{
			"order_id": &types.AttributeValueMemberS{Value: orderID},
			"id":       &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	if res.Item == nil {
		return nil, nil
	}
	var it models.OrderItem
	if err := attributevalue.UnmarshalMap(res.Item, &it); err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *DynamoRepository) ListOrderItems(ctx context.Context, orderID string) ([]models.OrderItem, error) {
	res, err := r.db.Query(ctx, &dynamodb.QueryInput{
		TableName:              &r.orderItemsTable,
		KeyConditionExpression: awsString("order_id = :oid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":oid": &types.AttributeValueMemberS{Value: orderID},
		},
	})
	if err != nil {
		return nil, err
	}
	var out []models.OrderItem
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *DynamoRepository) UpdateOrderItem(ctx context.Context, it *models.OrderItem) error {
	if it == nil {
		return errors.New("order item is nil")
	}
	item, err := attributevalue.MarshalMap(it)
	if err != nil {
		return err
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{TableName: &r.orderItemsTable, Item: item})
	return err
}

func (r *DynamoRepository) DeleteOrderItem(ctx context.Context, orderID, id string) error {
	_, err := r.db.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.orderItemsTable,
		Key: map[string]types.AttributeValue{
			"order_id": &types.AttributeValueMemberS{Value: orderID},
			"id":       &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

func awsString(s string) *string { return &s }
