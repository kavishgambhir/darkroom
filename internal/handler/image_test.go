package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/darkroom/pkg/service"
	"github.com/gojek/darkroom/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ImageHandlerTestSuite struct {
	suite.Suite
	deps        *service.Dependencies
	storage     *mockStorage
	manipulator *mockManipulator
}

func TestImageHandlerSuite(t *testing.T) {
	suite.Run(t, new(ImageHandlerTestSuite))
}

func (s *ImageHandlerTestSuite) SetupTest() {
	s.storage = &mockStorage{}
	s.manipulator = &mockManipulator{}
	s.deps = &service.Dependencies{Storage: s.storage, Manipulator: s.manipulator}
}

func (s *ImageHandlerTestSuite) TestImageHandler() {
	r, _ := http.NewRequest(http.MethodGet, "/image-valid", nil)
	rr := httptest.NewRecorder()

	s.storage.On("Get", mock.Anything, "/image-valid").Return([]byte("validData"), http.StatusOK, nil)

	ImageHandler(s.deps).ServeHTTP(rr, r)

	assert.Equal(s.T(), "validData", rr.Body.String())
	assert.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *ImageHandlerTestSuite) TestImageHandlerWithStorageGetError() {
	r, _ := http.NewRequest(http.MethodGet, "/image-invalid", nil)
	rr := httptest.NewRecorder()

	s.storage.On("Get", mock.Anything, "/image-invalid").Return([]byte(nil), http.StatusUnprocessableEntity, errors.New("error"))

	ImageHandler(s.deps).ServeHTTP(rr, r)

	assert.Equal(s.T(), "", rr.Body.String())
	assert.Equal(s.T(), http.StatusUnprocessableEntity, rr.Code)
}

func (s *ImageHandlerTestSuite) TestImageHandlerWithQueryParameters() {
	r, _ := http.NewRequest(http.MethodGet, "/image-valid?w=100&h=100", nil)
	rr := httptest.NewRecorder()

	params := make(map[string]string)
	params["w"] = "100"
	params["h"] = "100"
	s.storage.On("Get", mock.Anything, "/image-valid").Return([]byte("validData"), http.StatusOK, nil)
	s.manipulator.On("Process", mock.AnythingOfType("ProcessSpec")).Return([]byte("processedData"), nil)

	ImageHandler(s.deps).ServeHTTP(rr, r)

	assert.Equal(s.T(), "processedData", rr.Body.String())
	assert.Equal(s.T(), http.StatusOK, rr.Code)
}

func (s *ImageHandlerTestSuite) TestImageHandlerWithQueryParametersAndProcessingError() {
	r, _ := http.NewRequest(http.MethodGet, "/image-valid?w=100&h=100", nil)
	rr := httptest.NewRecorder()

	params := make(map[string]string)
	params["w"] = "100"
	params["h"] = "100"
	s.storage.On("Get", mock.Anything, "/image-valid").Return([]byte("validData"), http.StatusOK, nil)
	s.manipulator.On("Process", mock.AnythingOfType("ProcessSpec")).Return([]byte(nil), errors.New("error"))

	ImageHandler(s.deps).ServeHTTP(rr, r)

	assert.Equal(s.T(), "", rr.Body.String())
	assert.Equal(s.T(), http.StatusUnprocessableEntity, rr.Code)
}

type mockManipulator struct {
	mock.Mock
}

func (m *mockManipulator) Process(spec service.ProcessSpec) ([]byte, error) {
	args := m.Called(spec)
	return args.Get(0).([]byte), args.Error(1)
}

type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Get(ctx context.Context, path string) storage.IResponse {
	args := m.Called(ctx, path)
	return storage.NewResponse(args[0].([]byte), args.Int(1), args.Error(2))
}
