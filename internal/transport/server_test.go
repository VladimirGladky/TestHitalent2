package transport

import (
	"TestHitalent2/internal/config"
	"TestHitalent2/internal/models"
	"TestHitalent2/internal/service/mocks"
	"TestHitalent2/internal/suberrors"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateDepartmentsHandler_Success(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	inputDept := &models.Department{
		Name: "Engineering",
	}
	expectedDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		CreatedAt: time.Now(),
	}

	srv.EXPECT().CreateDepartment(gomock.Any()).Return(expectedDept, nil).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	body, _ := json.Marshal(inputDept)
	req := httptest.NewRequest("POST", "/api/v1/departments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	CreateDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response models.Department
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.Equal(t, expectedDept.ID, response.ID)
	require.Equal(t, expectedDept.Name, response.Name)
}

func TestCreateDepartmentsHandler_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	cases := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mocks.MockOrganizationServiceInterface)
		expectedStatus int
		checkError     bool
	}{
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockOrganizationServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:           "empty body",
			requestBody:    "",
			mockSetup:      func(m *mocks.MockOrganizationServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:        "invalid department name",
			requestBody: `{"name": ""}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateDepartment(gomock.Any()).Return(nil, suberrors.ErrInvalidDepartmentName).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
			checkError:     true,
		},
		{
			name:        "parent not found",
			requestBody: `{"name": "Test", "parent_id": 999}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateDepartment(gomock.Any()).Return(nil, suberrors.ErrParentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
			checkError:     true,
		},
		{
			name:        "duplicate name",
			requestBody: `{"name": "Engineering"}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateDepartment(gomock.Any()).Return(nil, suberrors.ErrDepartmentNameExistsInParent).Times(1)
			},
			expectedStatus: http.StatusConflict,
			checkError:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			srv := mocks.NewMockOrganizationServiceInterface(ctl)
			tc.mockSetup(srv)

			server := NewOrganizationServer(cfg, srv, ctx)

			req := httptest.NewRequest("POST", "/api/v1/departments", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			CreateDepartmentsHandler(server)(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
			if tc.checkError {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Contains(t, response, "error")
			}
		})
	}
}

func TestCreateDepartmentsHandler_UnknownError(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	srv.EXPECT().CreateDepartment(gomock.Any()).Return(nil, errors.New("unknown database error")).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	inputDept := &models.Department{Name: "Test"}
	body, _ := json.Marshal(inputDept)
	req := httptest.NewRequest("POST", "/api/v1/departments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	CreateDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateEmployeesHandler_Success(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	inputEmp := &models.Employee{
		FullName: "John Doe",
		Position: "Engineer",
	}
	expectedEmp := &models.Employee{
		ID:           1,
		DepartmentID: 1,
		FullName:     "John Doe",
		Position:     "Engineer",
		CreatedAt:    time.Now(),
	}

	srv.EXPECT().CreateEmployee(gomock.Any(), "1").Return(expectedEmp, nil).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	body, _ := json.Marshal(inputEmp)
	req := httptest.NewRequest("POST", "/api/v1/departments/1/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	CreateEmployeesHandler(server)(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var response models.Employee
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.Equal(t, expectedEmp.ID, response.ID)
	require.Equal(t, expectedEmp.FullName, response.FullName)
	require.Equal(t, expectedEmp.Position, response.Position)
}

func TestCreateEmployeesHandler_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	cases := []struct {
		name           string
		requestBody    string
		pathID         string
		mockSetup      func(*mocks.MockOrganizationServiceInterface)
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			pathID:         "1",
			mockSetup:      func(m *mocks.MockOrganizationServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid id",
			requestBody: `{"full_name": "Test", "position": "Dev"}`,
			pathID:      "abc",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateEmployee(gomock.Any(), "abc").Return(nil, suberrors.ErrInvalidID).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "department not found",
			requestBody: `{"full_name": "Test", "position": "Dev"}`,
			pathID:      "999",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateEmployee(gomock.Any(), "999").Return(nil, suberrors.ErrDepartmentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "invalid full name",
			requestBody: `{"full_name": "", "position": "Dev"}`,
			pathID:      "1",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().CreateEmployee(gomock.Any(), "1").Return(nil, suberrors.ErrInvalidEmployeeFullName).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			srv := mocks.NewMockOrganizationServiceInterface(ctl)
			tc.mockSetup(srv)

			server := NewOrganizationServer(cfg, srv, ctx)

			req := httptest.NewRequest("POST", "/api/v1/departments/"+tc.pathID+"/employees", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tc.pathID)

			w := httptest.NewRecorder()

			CreateEmployeesHandler(server)(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestGetDepartmentsHandler_Success(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	expectedDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		CreatedAt: time.Now(),
	}

	srv.EXPECT().GetDepartment("1", "1", "true").Return(expectedDept, nil).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	req := httptest.NewRequest("GET", "/api/v1/departments/1?depth=1&include_employees=true", nil)
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	GetDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response models.Department
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.Equal(t, expectedDept.ID, response.ID)
	require.Equal(t, expectedDept.Name, response.Name)
}

func TestGetDepartmentsHandler_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	cases := []struct {
		name           string
		pathID         string
		queryParams    string
		mockSetup      func(*mocks.MockOrganizationServiceInterface)
		expectedStatus int
	}{
		{
			name:        "invalid id",
			pathID:      "abc",
			queryParams: "",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().GetDepartment("abc", "", "").Return(nil, suberrors.ErrInvalidID).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "department not found",
			pathID:      "999",
			queryParams: "",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().GetDepartment("999", "", "").Return(nil, suberrors.ErrDepartmentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "invalid depth",
			pathID:      "1",
			queryParams: "?depth=abc",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().GetDepartment("1", "abc", "").Return(nil, suberrors.ErrInvalidDepth).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			srv := mocks.NewMockOrganizationServiceInterface(ctl)
			tc.mockSetup(srv)

			server := NewOrganizationServer(cfg, srv, ctx)

			req := httptest.NewRequest("GET", "/api/v1/departments/"+tc.pathID+tc.queryParams, nil)
			req.SetPathValue("id", tc.pathID)

			w := httptest.NewRecorder()

			GetDepartmentsHandler(server)(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestDeleteDepartmentsHandler_Success(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	srv.EXPECT().DeleteDepartment("1", "cascade", "").Return(nil).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	req := httptest.NewRequest("DELETE", "/api/v1/departments/1?mode=cascade", nil)
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	DeleteDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteDepartmentsHandler_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	cases := []struct {
		name           string
		pathID         string
		queryParams    string
		mockSetup      func(*mocks.MockOrganizationServiceInterface)
		expectedStatus int
	}{
		{
			name:        "invalid mode",
			pathID:      "1",
			queryParams: "?mode=invalid",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().DeleteDepartment("1", "invalid", "").Return(suberrors.ErrInvalidMode).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "department not found",
			pathID:      "999",
			queryParams: "?mode=cascade",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().DeleteDepartment("999", "cascade", "").Return(suberrors.ErrDepartmentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "reassign to self",
			pathID:      "5",
			queryParams: "?mode=reassign&reassign_to_department_id=5",
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().DeleteDepartment("5", "reassign", "5").Return(suberrors.ErrReassignToSelf).Times(1)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			srv := mocks.NewMockOrganizationServiceInterface(ctl)
			tc.mockSetup(srv)

			server := NewOrganizationServer(cfg, srv, ctx)

			req := httptest.NewRequest("DELETE", "/api/v1/departments/"+tc.pathID+tc.queryParams, nil)
			req.SetPathValue("id", tc.pathID)

			w := httptest.NewRecorder()

			DeleteDepartmentsHandler(server)(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestPatchDepartmentsHandler_Success(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)

	updateReq := &models.UpdateDepartmentRequest{
		Name: func() *string { s := "Updated Engineering"; return &s }(),
	}
	expectedDept := &models.Department{
		ID:        1,
		Name:      "Updated Engineering",
		CreatedAt: time.Now(),
	}

	srv.EXPECT().PatchDepartment("1", gomock.Any()).Return(expectedDept, nil).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest("PATCH", "/api/v1/departments/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	PatchDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response models.Department
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	require.Equal(t, expectedDept.ID, response.ID)
	require.Equal(t, expectedDept.Name, response.Name)
}

func TestPatchDepartmentsHandler_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	cases := []struct {
		name           string
		pathID         string
		requestBody    string
		mockSetup      func(*mocks.MockOrganizationServiceInterface)
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			pathID:         "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockOrganizationServiceInterface) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid id",
			pathID:      "abc",
			requestBody: `{"name": "Test"}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("abc", gomock.Any()).Return(nil, suberrors.ErrInvalidID).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "department not found",
			pathID:      "999",
			requestBody: `{"name": "Test"}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("999", gomock.Any()).Return(nil, suberrors.ErrDepartmentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "invalid department name",
			pathID:      "1",
			requestBody: `{"name": ""}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, suberrors.ErrInvalidDepartmentName).Times(1)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "parent not found",
			pathID:      "1",
			requestBody: `{"parent_id": 999}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, suberrors.ErrParentNotFound).Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "department own parent",
			pathID:      "1",
			requestBody: `{"parent_id": 1}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, suberrors.ErrDepartmentOwnParent).Times(1)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "department subtree",
			pathID:      "1",
			requestBody: `{"parent_id": 5}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, suberrors.ErrDepartmentSubtree).Times(1)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "duplicate name in parent",
			pathID:      "1",
			requestBody: `{"name": "Engineering"}`,
			mockSetup: func(m *mocks.MockOrganizationServiceInterface) {
				m.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, suberrors.ErrDepartmentNameExistsInParent).Times(1)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			srv := mocks.NewMockOrganizationServiceInterface(ctl)
			tc.mockSetup(srv)

			server := NewOrganizationServer(cfg, srv, ctx)

			req := httptest.NewRequest("PATCH", "/api/v1/departments/"+tc.pathID, bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("id", tc.pathID)

			w := httptest.NewRecorder()

			PatchDepartmentsHandler(server)(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestCreateEmployeesHandler_UnknownError(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)
	srv.EXPECT().CreateEmployee(gomock.Any(), "1").Return(nil, errors.New("unknown database error")).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	body, _ := json.Marshal(models.Employee{
		FullName:     "John Doe",
		Position:     "Engineer",
		DepartmentID: 1,
	})

	req := httptest.NewRequest("POST", "/api/v1/departments/1/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	CreateEmployeesHandler(server)(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetDepartmentsHandler_UnknownError(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)
	srv.EXPECT().GetDepartment("1", "", "").Return(nil, errors.New("unknown database error")).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	req := httptest.NewRequest("GET", "/api/v1/departments/1", nil)
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	GetDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPatchDepartmentsHandler_UnknownError(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)
	srv.EXPECT().PatchDepartment("1", gomock.Any()).Return(nil, errors.New("unknown database error")).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	body := `{"name": "Updated Department"}`
	req := httptest.NewRequest("PATCH", "/api/v1/departments/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	PatchDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteDepartmentsHandler_UnknownError(t *testing.T) {
	ctx := context.Background()
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	cfg := &config.Config{
		Host: "localhost",
		Port: "4053",
	}

	srv := mocks.NewMockOrganizationServiceInterface(ctl)
	srv.EXPECT().DeleteDepartment("1", "cascade", "").Return(errors.New("unknown database error")).Times(1)

	server := NewOrganizationServer(cfg, srv, ctx)

	req := httptest.NewRequest("DELETE", "/api/v1/departments/1?mode=cascade", nil)
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	DeleteDepartmentsHandler(server)(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
