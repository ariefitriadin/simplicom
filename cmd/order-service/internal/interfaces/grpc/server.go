package grpc

import (
	"context"
	"math/big"
	"time"

	pgrepo "github.com/ariefitriadin/simplicom/cmd/order-service/internal/persistence/postgres/repositories"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	proto "github.com/ariefitriadin/simplicom/cmd/order-service/proto"
	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
)

type OrderServer struct {
	queries                               *pgrepo.Queries
	db                                    *pgxpool.Pool
	proto.UnimplementedOrderServiceServer // Embed the unimplemented server
}

func NewServer(queries *pgrepo.Queries, db *pgxpool.Pool) proto.OrderServiceServer {
	return &OrderServer{queries: queries, db: db}
}

func (s *OrderServer) CreateOrder(ctx context.Context, req *proto.CreateOrderRequest) (*proto.CreateOrderResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer tx.Rollback(ctx)

	// Generate ULID for order ID
	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(uint64(t.UnixNano()))), 0)
	orderID := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	orderPayload := req.GetOrder()
	order, err := s.queries.WithTx(tx).InsertOrder(ctx, pgrepo.InsertOrderParams{
		ID:         uuid.MustParse(orderID),
		CustomerID: orderPayload.CustomerId,
		OrderDate:  pgtype.Timestamp{Time: orderPayload.OrderDate.AsTime(), Valid: true},
		Status:     orderPayload.Status,
		Total:      pgtype.Numeric{Int: big.NewInt(int64(orderPayload.Total)), Valid: true},
	})

	var orderItems pgrepo.InsertOrderItemParams
	for _, item := range orderPayload.Items {
		orderItems.Column1 = append(orderItems.Column1, uuid.MustParse(orderID))
		orderItems.Column2 = append(orderItems.Column2, item.ProductId)
		orderItems.Column3 = append(orderItems.Column3, item.Quantity)
		orderItems.Column4 = append(orderItems.Column4, pgtype.Numeric{Int: big.NewInt(int64(item.Price)), Valid: true})
	}

	// Execute bulk insert using UNNEST
	err = s.queries.WithTx(tx).InsertOrderItem(ctx, orderItems)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	orderTotal, err := order.Total.Float64Value()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &proto.CreateOrderResponse{
		Order: &proto.Order{
			Id:         order.ID.String(),
			CustomerId: order.CustomerID,
			OrderDate:  timestamppb.New(order.OrderDate.Time),
			Status:     order.Status,
			Total:      orderTotal.Float64,
		},
	}, nil
}

// UpdateOrder implements the UpdateOrder RPC method
func (s *OrderServer) UpdateOrder(ctx context.Context, req *proto.UpdateOrderRequest) (*proto.UpdateOrderResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer tx.Rollback(ctx)

	orderPayload := req.GetOrder()

	// Check if the order exists
	existingOrder, err := s.queries.WithTx(tx).GetOrder(ctx, uuid.MustParse(orderPayload.Id))
	if err != nil {
		if err.Error() == apperrors.ErrNoRows.Error() {
			return nil, apperrors.New("order not found")
		}
		return nil, apperrors.Wrap(err)
	}

	// Update the order
	order, err := s.queries.WithTx(tx).UpdateOrder(ctx, pgrepo.UpdateOrderParams{
		ID:     existingOrder.ID,
		Status: orderPayload.Status,
		Total:  pgtype.Numeric{Int: big.NewInt(int64(orderPayload.Total)), Valid: true},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var orderItems pgrepo.UpdateOrderItemsParams
	for _, item := range orderPayload.Items {
		orderItems.Column1 = append(orderItems.Column1, uuid.MustParse(item.Id))
		orderItems.Column2 = append(orderItems.Column2, existingOrder.ID)
		orderItems.Column3 = append(orderItems.Column3, item.ProductId)
		orderItems.Column4 = append(orderItems.Column4, item.Quantity)
		orderItems.Column5 = append(orderItems.Column5, pgtype.Numeric{Int: big.NewInt(int64(item.Price)), Valid: true})
	}

	// Execute bulk update using UNNEST
	err = s.queries.WithTx(tx).UpdateOrderItems(ctx, orderItems)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	orderTotal, err := order.Total.Float64Value()
	if err != nil {
		return nil, err
	}

	return &proto.UpdateOrderResponse{
		Order: &proto.Order{
			Id:        order.ID.String(),
			OrderDate: timestamppb.New(order.OrderDate.Time),
			Status:    order.Status,
			Total:     orderTotal.Float64,
		},
	}, nil
}
