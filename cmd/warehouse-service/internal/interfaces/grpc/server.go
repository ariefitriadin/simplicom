package grpc

import (
	"context"
	"time"

	pgrepo "github.com/ariefitriadin/simplicom/cmd/warehouse-service/internal/persistence/postgres/repositories"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	proto "github.com/ariefitriadin/simplicom/cmd/warehouse-service/proto"
	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
)

type WarehouseServer struct {
	queries                                   *pgrepo.Queries
	db                                        *pgxpool.Pool
	proto.UnimplementedWarehouseServiceServer // Embed the unimplemented server
}

func NewServer(queries *pgrepo.Queries, db *pgxpool.Pool) proto.WarehouseServiceServer {
	return &WarehouseServer{queries: queries, db: db}
}

func (s *WarehouseServer) CreateWarehouse(ctx context.Context, req *proto.CreateWarehouseRequest) (*proto.Warehouse, error) {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.New("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	warehouse, err := s.queries.WithTx(tx).CreateWarehouse(ctx, pgrepo.CreateWarehouseParams{
		Name:         req.Name,
		LocationID:   uuid.MustParse(req.LocationId),
		LocationName: req.Location,
		Capacity:     int32(req.Capacity),
	})
	if err != nil {
		return nil, apperrors.New("failed to create warehouse")
	}

	return &proto.Warehouse{
		Id:       warehouse.ID.String(),
		Name:     warehouse.Name,
		Location: warehouse.LocationName,
		Capacity: int32(warehouse.Capacity),
	}, nil

}

func (s *WarehouseServer) GetWarehouse(ctx context.Context, req *proto.GetWarehouseRequest) (*proto.Warehouse, error) {
	// Implement get warehouse logic
	return nil, apperrors.New("method GetWarehouse not implemented")
}

func (s *WarehouseServer) ListWarehouses(ctx context.Context, req *proto.ListWarehousesRequest) (*proto.ListWarehousesResponse, error) {
	// Implement list warehouses logic
	return nil, apperrors.New("method ListWarehouses not implemented")
}

func (s *WarehouseServer) UpdateWarehouse(ctx context.Context, req *proto.UpdateWarehouseRequest) (*proto.Warehouse, error) {
	// Implement update warehouse logic
	return nil, apperrors.New("method UpdateWarehouse not implemented")
}

func (s *WarehouseServer) DeleteWarehouse(ctx context.Context, req *proto.DeleteWarehouseRequest) (*proto.DeleteWarehouseResponse, error) {
	// Implement delete warehouse logic
	return nil, apperrors.New("method DeleteWarehouse not implemented")
}

func (s *WarehouseServer) AddStock(ctx context.Context, req *proto.AddStockRequest) (*proto.StockResponse, error) {
	// Implement add stock logic
	return nil, apperrors.New("method AddStock not implemented")
}

func (s *WarehouseServer) GetStock(ctx context.Context, req *proto.GetStockRequest) (*proto.StockResponse, error) {
	// Implement get stock logic
	return nil, apperrors.New("method GetStock not implemented")
}

func (s *WarehouseServer) ReserveStock(ctx context.Context, req *proto.ReserveStockRequest) (*proto.ReservationResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.New("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	resTime, err := s.queries.WithTx(tx).ReserveStock(ctx, pgrepo.ReserveStockParams{
		WarehouseID: uuid.MustParse(req.WarehouseId),
		ProductID:   uuid.MustParse(req.ProductId),
		OrderID:     uuid.MustParse(req.OrderId),
		Column1:     req.Quantity,
		Column5:     pgtype.Interval{Microseconds: int64(time.Hour / time.Microsecond), Days: 0, Months: 0, Valid: true}, //set expired to 1 hour
	})
	if err != nil {
		return nil, apperrors.New("failed to reserve stock")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.New("failed to commit transaction")
	}

	return &proto.ReservationResponse{
		ExpiresAt: timestamppb.New(resTime.Time),
	}, nil
}

func (s *WarehouseServer) ConfirmReservation(ctx context.Context, req *proto.ConfirmReservationRequest) (*proto.ConfirmationResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.New("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	cr, err := s.queries.WithTx(tx).ConfirmReservation(ctx, pgrepo.ConfirmReservationParams{
		ID:      uuid.MustParse(req.ReservationId),
		OrderID: uuid.MustParse(req.OrderId),
	})
	if err != nil {
		return nil, apperrors.New("failed to confirm reservation")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, apperrors.New("failed to commit transaction")
	}

	return &proto.ConfirmationResponse{
		Success: true,
	}, nil
}

func (s *WarehouseServer) CancelReservation(ctx context.Context, req *proto.CancelReservationRequest) (*proto.CancellationResponse, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, apperrors.New("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	_, err = s.queries.WithTx(tx).CancelExpiredReservations(ctx)
	if err != nil {
		return nil, apperrors.New("failed to cancel reservation")
	}

	return &proto.CancellationResponse{
		Success: true,
	}, nil
}

func (s *WarehouseServer) GetStockHistory(ctx context.Context, req *proto.GetStockHistoryRequest) (*proto.StockHistoryResponse, error) {
	// Implement get stock history logic
	return nil, apperrors.New("method GetStockHistory not implemented")
}
