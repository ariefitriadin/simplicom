package grpc

import (
	"context"
	"fmt"
	"math/big"

	pgrepo "github.com/ariefitriadin/simplicom/cmd/product-service/internal/persistence/postgres/repositories"
	apperrors "github.com/ariefitriadin/simplicom/pkg/errors"
	"github.com/ariefitriadin/simplicom/pkg/logger"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	proto "github.com/ariefitriadin/simplicom/cmd/product-service/proto"
)

type ProductServer struct {
	queries                                 *pgrepo.Queries
	db                                      *pgxpool.Pool
	proto.UnimplementedProductServiceServer // Embed the unimplemented server
}

func NewServer(queries *pgrepo.Queries, db *pgxpool.Pool) proto.ProductServiceServer {
	return &ProductServer{queries: queries, db: db}
}

func (s *ProductServer) CreateProduct(ctx context.Context, request *proto.CreateProductRequest) (*proto.CreateProductResponse, error) {

	// insert one product
	product, err := s.queries.InsertProduct(ctx, pgrepo.InsertProductParams{
		Name:        request.Name,
		Description: pgtype.Text{String: request.Description, Valid: true},
		Price:       pgtype.Numeric{Int: big.NewInt(int64(request.Price)), Valid: true},
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	prodPrice, err := product.Price.Float64Value()
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	return &proto.CreateProductResponse{Product: &proto.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description.String,
		Price:       prodPrice.Float64,
		StockLevel:  0,
		WarehouseId: 0,
	}}, nil
}

func (s *ProductServer) GetProducts(ctx context.Context, request *proto.GetProductsRequest) (*proto.GetProductsResponse, error) {

	products, err := s.queries.GetAllProductsWithStock(ctx, pgrepo.GetAllProductsWithStockParams{
		Offset: request.Offset,
		Limit:  request.Limit,
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	var pbProducts []*proto.Product
	for _, product := range products {
		prodPrice, err := product.Price.Float64Value()
		if err != nil {
			logger.Error(ctx, err.Error())
			return nil, apperrors.Wrap(err)
		}
		pbProducts = append(pbProducts, &proto.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       prodPrice.Float64,
			StockLevel:  int32(product.StockLevel.Int32),
			WarehouseId: int32(product.WarehouseID.Int32),
		})
	}

	return &proto.GetProductsResponse{
		Products: pbProducts,
		Total:    int32(len(products)),
	}, nil
}

func (s *ProductServer) UpdateProductStock(ctx context.Context, request *proto.UpdateProductStockRequest) (*proto.UpdateProductStockResponse, error) {

	//check product availability
	product, err := s.queries.GetProductByID(ctx, request.ProductId)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(fmt.Errorf("product not found"))
	}

	err = s.queries.UpdateProductStock(ctx, pgrepo.UpdateProductStockParams{
		ProductID:     request.ProductId,
		StockLevel:    pgtype.Int4{Int32: request.StockLevel, Valid: true},
		WarehouseID:   request.WarehouseId,
		WarehouseID_2: request.WhereWhouseId,
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	prodPrice, err := product.Price.Float64Value()
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}
	return &proto.UpdateProductStockResponse{
		Product: &proto.Product{
			Id:          request.ProductId,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       prodPrice.Float64,
			WarehouseId: request.WarehouseId,
			StockLevel:  request.StockLevel,
		},
	}, nil
}

func (s *ProductServer) InsertProductStock(ctx context.Context, request *proto.InsertProductStockRequest) (*proto.InsertProductStockResponse, error) {

	_, err := s.queries.InsertProductStock(ctx, pgrepo.InsertProductStockParams{
		ProductID:   request.ProductId,
		StockLevel:  pgtype.Int4{Int32: request.StockLevel, Valid: true},
		WarehouseID: request.WarehouseId,
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	product, err := s.queries.GetProductByID(ctx, request.ProductId)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	prodPrice, err := product.Price.Float64Value()
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, apperrors.Wrap(err)
	}

	return &proto.InsertProductStockResponse{
		Product: &proto.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description.String,
			Price:       prodPrice.Float64,
			WarehouseId: request.WarehouseId,
			StockLevel:  request.StockLevel,
		},
	}, nil
}
