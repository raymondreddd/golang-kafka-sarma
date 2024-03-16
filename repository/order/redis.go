package order

import (
	"context"
	"errors"
	"fmt"

	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/raymondreddd/golnag/model"
)

type RedisRepo struct {
	Client *redis.Client
}

var ErrNotExist = errors.New("order does not exist")

func orderIDKey(id uint64) string {
	return fmt.Sprintf("Order:%d", id)
}

// Create method
func (r *RedisRepo) Create(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("error in encoding order: %w", err)
	}

	// generate key
	key := orderIDKey(order.OrderID)

	// transaction instancefor 2 ops
	tx := r.Client.TxPipeline()

	// save to redis
	res := tx.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		// discard complete trans for
		tx.Discard()
		return fmt.Errorf("error in setting in redis: %w", err)
	}

	if err := tx.SAdd(ctx, "orders", key).Err(); err != nil {
		// discard complete trans for
		tx.Discard()
		return fmt.Errorf("error in adding to orders ser: %w", err)
	}

	// Execute both transaction
	if _, err := tx.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	// save to setall
	return nil
}

// Get
func (r *RedisRepo) FindById(ctx context.Context, id uint64) (model.Order, error) {
	// convert to string
	key := orderIDKey(id)

	// save to redis
	resJson, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	// now convert the JSON result into
	var order model.Order

	err = json.Unmarshal([]byte(resJson), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("error in decoding order json %w", err)
	}

	return order, nil
}

func (r *RedisRepo) DeleteById(ctx context.Context, id uint64) error {
	// convert to string
	key := orderIDKey(id)

	// transaction instancefor 2 ops
	tx := r.Client.TxPipeline()

	// save to redis
	err := tx.Del(ctx, key).Err()

	if errors.Is(err, redis.Nil) {
		tx.Discard()
		return ErrNotExist
	} else if err != nil {
		tx.Discard()
		return fmt.Errorf("get order: %w", err)
	}

	if err := tx.SRem(ctx, "orders", key).Err(); err != nil {
		tx.Discard()
		return fmt.Errorf("error in removing from order set: %w", err)
	}

	return nil
}

type FindAllPage struct {
	Size   uint
	Offset uint
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAllPage(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", uint64(page.Offset), "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("error in fetching orders IDs %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
				Orders: []model.Order{},
			},
			nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("error in fetching orderss %w", err)
	}

	// create orders slice of type Order
	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)

		var order model.Order

		if err := json.Unmarshal([]byte(x), &order); err != nil {
			return FindResult{}, fmt.Errorf("error in decoding orders %w", err)
		}

		orders[i] = order
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
