package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/meetoria/meetoria/backend/internal/customer"
	customerservice "github.com/meetoria/meetoria/backend/internal/customer/service"
	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

type mockCustomerRepo struct {
	mock.Mock
}

func (m *mockCustomerRepo) Create(ctx context.Context, c *customer.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCustomerRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*customer.Customer, error) {
	args := m.Called(ctx, orgID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *mockCustomerRepo) FindByPhoneOrEmail(ctx context.Context, orgID uuid.UUID, phone, email string) (*customer.Customer, error) {
	args := m.Called(ctx, orgID, phone, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*customer.Customer), args.Error(1)
}

func (m *mockCustomerRepo) Update(ctx context.Context, c *customer.Customer) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *mockCustomerRepo) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	args := m.Called(ctx, orgID, id)
	return args.Error(0)
}

func (m *mockCustomerRepo) List(ctx context.Context, orgID uuid.UUID, offset, limit int, search string) ([]customer.Customer, int64, error) {
	args := m.Called(ctx, orgID, offset, limit, search)
	return args.Get(0).([]customer.Customer), args.Get(1).(int64), args.Error(2)
}

func (m *mockCustomerRepo) GetBookingStats(ctx context.Context, orgID uuid.UUID, customerIDs []uuid.UUID) (map[uuid.UUID]customer.BookingStats, error) {
	args := m.Called(ctx, orgID, customerIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]customer.BookingStats), args.Error(1)
}

func TestCreateCustomer(t *testing.T) {
	repo := new(mockCustomerRepo)
	svc := customerservice.NewService(repo, nil, nil)
	orgID := uuid.New()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*customer.Customer")).Return(nil)

	req := customer.CreateCustomerRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "+37060000000",
	}

	result, err := svc.Create(context.Background(), orgID, req)
	assert.NoError(t, err)
	assert.Equal(t, "John", result.FirstName)
	assert.Equal(t, orgID, result.OrganizationID)
}

func TestGetCustomerNotFound(t *testing.T) {
	repo := new(mockCustomerRepo)
	svc := customerservice.NewService(repo, nil, nil)
	orgID := uuid.New()
	customerID := uuid.New()

	repo.On("GetByID", mock.Anything, orgID, customerID).Return(nil, gorm.ErrRecordNotFound)

	result, err := svc.GetByID(context.Background(), orgID, customerID)
	assert.Nil(t, result)
	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "NOT_FOUND", appErr.Code)
}

func TestListCustomers(t *testing.T) {
	repo := new(mockCustomerRepo)
	svc := customerservice.NewService(repo, nil, nil)
	orgID := uuid.New()

	customers := []customer.Customer{
		{OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID}, FirstName: "A", LastName: "B"},
	}
	repo.On("List", mock.Anything, orgID, 0, 20, "").Return(customers, int64(1), nil)
	repo.On("GetBookingStats", mock.Anything, orgID, mock.Anything).Return(map[uuid.UUID]customer.BookingStats{}, nil)

	result, err := svc.List(context.Background(), orgID, commonmodel.PaginationParams{Page: 1, Limit: 20}, "")
	assert.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, int64(1), result.Total)
}
