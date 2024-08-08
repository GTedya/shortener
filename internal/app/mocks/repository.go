// Code generated by MockGen. DO NOT EDIT.
// Source: repository.go

// Package mock_repository is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/GTedya/shortener/internal/app/models"
	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockRepository) Check(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockRepositoryMockRecorder) Check(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockRepository)(nil).Check), ctx)
}

// Close mocks base method.
func (m *MockRepository) Close(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRepositoryMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRepository)(nil).Close), arg0)
}

// DeleteUrls mocks base method.
func (m *MockRepository) DeleteUrls(ctx context.Context, urls []models.ShortURL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUrls", ctx, urls)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUrls indicates an expected call of DeleteUrls.
func (mr *MockRepositoryMockRecorder) DeleteUrls(ctx, urls interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUrls", reflect.TypeOf((*MockRepository)(nil).DeleteUrls), ctx, urls)
}

// GetByID mocks base method.
func (m *MockRepository) GetByID(ctx context.Context, id string) (models.ShortURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(models.ShortURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockRepository)(nil).GetByID), ctx, id)
}

// GetUsersAndUrlsCount mocks base method.
func (m *MockRepository) GetUsersAndUrlsCount(ctx context.Context) (int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsersAndUrlsCount", ctx)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetUsersAndUrlsCount indicates an expected call of GetUsersAndUrlsCount.
func (mr *MockRepositoryMockRecorder) GetUsersAndUrlsCount(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsersAndUrlsCount", reflect.TypeOf((*MockRepository)(nil).GetUsersAndUrlsCount), ctx)
}

// GetUsersUrls mocks base method.
func (m *MockRepository) GetUsersUrls(ctx context.Context, userID string) ([]models.ShortURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsersUrls", ctx, userID)
	ret0, _ := ret[0].([]models.ShortURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsersUrls indicates an expected call of GetUsersUrls.
func (mr *MockRepositoryMockRecorder) GetUsersUrls(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsersUrls", reflect.TypeOf((*MockRepository)(nil).GetUsersUrls), ctx, userID)
}

// Save mocks base method.
func (m *MockRepository) Save(ctx context.Context, shortURL models.ShortURL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, shortURL)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockRepositoryMockRecorder) Save(ctx, shortURL interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockRepository)(nil).Save), ctx, shortURL)
}

// SaveBatch mocks base method.
func (m *MockRepository) SaveBatch(ctx context.Context, batch []models.ShortURL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBatch", ctx, batch)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveBatch indicates an expected call of SaveBatch.
func (mr *MockRepositoryMockRecorder) SaveBatch(ctx, batch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBatch", reflect.TypeOf((*MockRepository)(nil).SaveBatch), ctx, batch)
}

// ShortenByURL mocks base method.
func (m *MockRepository) ShortenByURL(ctx context.Context, url string) (models.ShortURL, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShortenByURL", ctx, url)
	ret0, _ := ret[0].(models.ShortURL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShortenByURL indicates an expected call of ShortenByURL.
func (mr *MockRepositoryMockRecorder) ShortenByURL(ctx, url interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortenByURL", reflect.TypeOf((*MockRepository)(nil).ShortenByURL), ctx, url)
}
