package service

import (
	"TestHitalent2/internal/models"
	"TestHitalent2/internal/repository/mocks"
	"TestHitalent2/internal/suberrors"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestOrganizationService_CreateDepartment_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	dept := &models.Department{
		Name: "Engineering",
	}
	expResp := &models.Department{
		ID:        1,
		Name:      "Engineering",
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	repo.EXPECT().DepartmentNameExistsInParent("Engineering", (*int)(nil), 0).Return(false, nil).Times(1)
	repo.EXPECT().CreateDepartment(dept).Return(expResp, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.CreateDepartment(dept)

	require.NoError(t, err)
	require.Equal(t, expResp, department)
}

func TestOrganizationService_CreateDepartment_WithParent_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	parentID := 1
	dept := &models.Department{
		Name:     "Backend Team",
		ParentID: &parentID,
	}
	expResp := &models.Department{
		ID:        2,
		Name:      "Backend Team",
		ParentID:  &parentID,
		CreatedAt: time.Now(),
	}

	repo.EXPECT().DepartmentExists(parentID).Return(true, nil).Times(1)
	repo.EXPECT().DepartmentNameExistsInParent("Backend Team", &parentID, 0).Return(false, nil).Times(1)
	repo.EXPECT().CreateDepartment(dept).Return(expResp, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.CreateDepartment(dept)

	require.NoError(t, err)
	require.Equal(t, expResp, department)
}

func TestOrganizationService_CreateDepartment_Fail(t *testing.T) {
	cases := []struct {
		name        string
		department  *models.Department
		setupMock   func(*mocks.MockOrganizationRepositoryInterface)
		expectedErr error
	}{
		{
			name:        "nil department",
			department:  nil,
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrNilDepartment,
		},
		{
			name: "empty name after trim",
			department: &models.Department{
				Name: "   ",
			},
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidDepartmentName,
		},
		{
			name: "name too long",
			department: &models.Department{
				Name: strings.Repeat("a", 201),
			},
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidDepartmentName,
		},
		{
			name: "parent not found",
			department: &models.Department{
				Name:     "Test",
				ParentID: func() *int { i := 999; return &i }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				m.EXPECT().DepartmentExists(999).Return(false, nil).Times(1)
			},
			expectedErr: suberrors.ErrParentNotFound,
		},
		{
			name: "duplicate name in parent",
			department: &models.Department{
				Name:     "Test",
				ParentID: func() *int { i := 1; return &i }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				m.EXPECT().DepartmentExists(1).Return(true, nil).Times(1)
				m.EXPECT().DepartmentNameExistsInParent("Test", gomock.Any(), 0).Return(true, nil).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentNameExistsInParent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
			tc.setupMock(repo)

			srv := NewOrganizationService(context.Background(), repo)
			department, err := srv.CreateDepartment(tc.department)

			require.Error(t, err)
			require.Nil(t, department)
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestOrganizationService_CreateEmployee_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	emp := &models.Employee{
		FullName: "John Doe",
		Position: "Engineer",
	}
	expResp := &models.Employee{
		ID:           1,
		DepartmentID: 1,
		FullName:     "John Doe",
		Position:     "Engineer",
		CreatedAt:    time.Now(),
	}

	repo.EXPECT().DepartmentExists(1).Return(true, nil).Times(1)
	repo.EXPECT().CreateEmployee(gomock.Any()).Return(expResp, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	employee, err := srv.CreateEmployee(emp, "1")

	require.NoError(t, err)
	require.Equal(t, expResp, employee)
}

func TestOrganizationService_CreateEmployee_Fail(t *testing.T) {
	cases := []struct {
		name        string
		employee    *models.Employee
		id          string
		setupMock   func(*mocks.MockOrganizationRepositoryInterface)
		expectedErr error
	}{
		{
			name:        "nil employee",
			employee:    nil,
			id:          "1",
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrNilEmployee,
		},
		{
			name:        "invalid id",
			employee:    &models.Employee{FullName: "Test", Position: "Dev"},
			id:          "abc",
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidID,
		},
		{
			name:        "empty full name after trim",
			employee:    &models.Employee{FullName: "   ", Position: "Dev"},
			id:          "1",
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidEmployeeFullName,
		},
		{
			name:        "empty position after trim",
			employee:    &models.Employee{FullName: "Test", Position: "   "},
			id:          "1",
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidEmployeePosition,
		},
		{
			name:     "department not found",
			employee: &models.Employee{FullName: "Test", Position: "Dev"},
			id:       "999",
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				m.EXPECT().DepartmentExists(999).Return(false, nil).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
			tc.setupMock(repo)

			srv := NewOrganizationService(context.Background(), repo)
			employee, err := srv.CreateEmployee(tc.employee, tc.id)

			require.Error(t, err)
			require.Nil(t, employee)
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestOrganizationService_GetDepartment_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	expDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		CreatedAt: time.Now(),
	}

	repo.EXPECT().GetDepartment(1, true).Return(expDept, nil).Times(1)
	repo.EXPECT().GetChildrenByParentID(1, true).Return([]*models.Department{}, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.GetDepartment("1", "1", "true")

	require.NoError(t, err)
	require.Equal(t, expDept, department)
}

func TestOrganizationService_GetDepartment_Fail(t *testing.T) {
	cases := []struct {
		name             string
		id               string
		depth            string
		includeEmployees string
		setupMock        func(*mocks.MockOrganizationRepositoryInterface)
		expectedErr      error
	}{
		{
			name:             "invalid id",
			id:               "abc",
			depth:            "1",
			includeEmployees: "true",
			setupMock:        func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:      suberrors.ErrInvalidID,
		},
		{
			name:             "invalid depth",
			id:               "1",
			depth:            "abc",
			includeEmployees: "true",
			setupMock:        func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:      suberrors.ErrInvalidDepth,
		},
		{
			name:             "invalid include_employees",
			id:               "1",
			depth:            "1",
			includeEmployees: "maybe",
			setupMock:        func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:      suberrors.ErrInvalidIncludeEmployees,
		},
		{
			name:             "department not found",
			id:               "999",
			depth:            "1",
			includeEmployees: "true",
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				m.EXPECT().GetDepartment(999, true).Return(nil, gorm.ErrRecordNotFound).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
			tc.setupMock(repo)

			srv := NewOrganizationService(context.Background(), repo)
			department, err := srv.GetDepartment(tc.id, tc.depth, tc.includeEmployees)

			require.Error(t, err)
			require.Nil(t, department)
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestOrganizationService_GetDepartment_WithChildren_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	expDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		CreatedAt: time.Now(),
	}
	child1 := &models.Department{ID: 2, Name: "Backend"}
	child2 := &models.Department{ID: 3, Name: "Frontend"}

	repo.EXPECT().GetDepartment(1, true).Return(expDept, nil).Times(1)
	repo.EXPECT().GetChildrenByParentID(1, true).Return([]*models.Department{child1, child2}, nil).Times(1)
	repo.EXPECT().GetChildrenByParentID(2, true).Return([]*models.Department{}, nil).Times(1)
	repo.EXPECT().GetChildrenByParentID(3, true).Return([]*models.Department{}, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.GetDepartment("1", "2", "true")

	require.NoError(t, err)
	require.Equal(t, expDept, department)
	require.Len(t, department.Children, 2)
}

func TestOrganizationService_CreateDepartment_DatabaseError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	parentID := 1
	dept := &models.Department{
		Name:     "Backend Team",
		ParentID: &parentID,
	}

	repo.EXPECT().DepartmentExists(parentID).Return(false, errors.New("database error")).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.CreateDepartment(dept)

	require.Error(t, err)
	require.Nil(t, department)
	require.Contains(t, err.Error(), "database error")
}

func TestOrganizationService_CreateEmployee_DatabaseError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	emp := &models.Employee{
		FullName: "John Doe",
		Position: "Engineer",
	}

	repo.EXPECT().DepartmentExists(1).Return(false, errors.New("database error")).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	employee, err := srv.CreateEmployee(emp, "1")

	require.Error(t, err)
	require.Nil(t, employee)
	require.Contains(t, err.Error(), "database error")
}

func TestOrganizationService_GetDepartment_LoadChildrenError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
	expDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		CreatedAt: time.Now(),
	}

	repo.EXPECT().GetDepartment(1, true).Return(expDept, nil).Times(1)
	repo.EXPECT().GetChildrenByParentID(1, true).Return(nil, errors.New("database error")).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.GetDepartment("1", "1", "true")

	require.Error(t, err)
	require.Nil(t, department)
	require.Contains(t, err.Error(), "database error")
}

func TestOrganizationService_PatchDepartment_DatabaseError(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)

	newName := "Updated Engineering"
	updateReq := &models.UpdateDepartmentRequest{
		Name: &newName,
	}

	currentDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	repo.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
	repo.EXPECT().DepartmentNameExistsInParent("Updated Engineering", (*int)(nil), 1).Return(false, errors.New("database error")).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.PatchDepartment("1", updateReq)

	require.Error(t, err)
	require.Nil(t, department)
	require.Contains(t, err.Error(), "database error")
}

func TestOrganizationService_DeleteDepartment_Cascade_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)

	repo.EXPECT().DeleteDepartmentCascade(1).Return(nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	err := srv.DeleteDepartment("1", "cascade", "")

	require.NoError(t, err)
}

func TestOrganizationService_DeleteDepartment_Reassign_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)

	repo.EXPECT().DeleteDepartmentReassign(1, 2).Return(nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	err := srv.DeleteDepartment("1", "reassign", "2")

	require.NoError(t, err)
}

func TestOrganizationService_DeleteDepartment_Fail(t *testing.T) {
	cases := []struct {
		name                   string
		id                     string
		mode                   string
		reassignToDepartmentId string
		setupMock              func(*mocks.MockOrganizationRepositoryInterface)
		expectedErr            error
	}{
		{
			name:                   "invalid id",
			id:                     "abc",
			mode:                   "cascade",
			reassignToDepartmentId: "",
			setupMock:              func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:            suberrors.ErrInvalidID,
		},
		{
			name:                   "empty mode",
			id:                     "1",
			mode:                   "",
			reassignToDepartmentId: "",
			setupMock:              func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:            suberrors.ErrEmptyMode,
		},
		{
			name:                   "invalid mode",
			id:                     "1",
			mode:                   "delete_all",
			reassignToDepartmentId: "",
			setupMock:              func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:            suberrors.ErrInvalidMode,
		},
		{
			name:                   "reassign to self",
			id:                     "5",
			mode:                   "reassign",
			reassignToDepartmentId: "5",
			setupMock:              func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr:            suberrors.ErrReassignToSelf,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
			tc.setupMock(repo)

			srv := NewOrganizationService(context.Background(), repo)
			err := srv.DeleteDepartment(tc.id, tc.mode, tc.reassignToDepartmentId)

			require.Error(t, err)
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestOrganizationService_PatchDepartment_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)

	newName := "Updated Engineering"
	updateReq := &models.UpdateDepartmentRequest{
		Name: &newName,
	}

	currentDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	expectedDept := &models.Department{
		ID:        1,
		Name:      "Updated Engineering",
		ParentID:  nil,
		CreatedAt: currentDept.CreatedAt,
	}

	repo.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
	repo.EXPECT().DepartmentNameExistsInParent("Updated Engineering", (*int)(nil), 1).Return(false, nil).Times(1)
	repo.EXPECT().PatchDepartment(gomock.Any()).Return(expectedDept, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.PatchDepartment("1", updateReq)

	require.NoError(t, err)
	require.Equal(t, expectedDept, department)
}

func TestOrganizationService_PatchDepartment_ChangeParent_Success(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	repo := mocks.NewMockOrganizationRepositoryInterface(ctl)

	newParentID := 2
	updateReq := &models.UpdateDepartmentRequest{
		ParentID: &newParentID,
	}

	currentDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	expectedDept := &models.Department{
		ID:        1,
		Name:      "Engineering",
		ParentID:  &newParentID,
		CreatedAt: currentDept.CreatedAt,
	}

	repo.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
	repo.EXPECT().DepartmentExists(newParentID).Return(true, nil).Times(1)
	repo.EXPECT().IsChildOf(newParentID, 1).Return(false, nil).Times(1)
	repo.EXPECT().PatchDepartment(gomock.Any()).Return(expectedDept, nil).Times(1)

	srv := NewOrganizationService(context.Background(), repo)
	department, err := srv.PatchDepartment("1", updateReq)

	require.NoError(t, err)
	require.Equal(t, expectedDept, department)
}

func TestOrganizationService_PatchDepartment_Fail(t *testing.T) {
	cases := []struct {
		name        string
		id          string
		updateReq   *models.UpdateDepartmentRequest
		setupMock   func(*mocks.MockOrganizationRepositoryInterface)
		expectedErr error
	}{
		{
			name:        "invalid id",
			id:          "abc",
			updateReq:   &models.UpdateDepartmentRequest{Name: func() *string { s := "Test"; return &s }()},
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrInvalidID,
		},
		{
			name:        "nil update request",
			id:          "1",
			updateReq:   nil,
			setupMock:   func(m *mocks.MockOrganizationRepositoryInterface) {},
			expectedErr: suberrors.ErrNilDepartmentUpdate,
		},
		{
			name:      "department not found",
			id:        "999",
			updateReq: &models.UpdateDepartmentRequest{Name: func() *string { s := "Test"; return &s }()},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				m.EXPECT().GetDepartment(999, false).Return(nil, gorm.ErrRecordNotFound).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentNotFound,
		},
		{
			name:      "invalid department name",
			id:        "1",
			updateReq: &models.UpdateDepartmentRequest{Name: func() *string { s := ""; return &s }()},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				currentDept := &models.Department{ID: 1, Name: "Test"}
				m.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
			},
			expectedErr: suberrors.ErrInvalidDepartmentName,
		},
		{
			name: "parent not found",
			id:   "1",
			updateReq: &models.UpdateDepartmentRequest{
				ParentID: func() *int { i := 999; return &i }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				currentDept := &models.Department{ID: 1, Name: "Test"}
				m.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
				m.EXPECT().DepartmentExists(999).Return(false, nil).Times(1)
			},
			expectedErr: suberrors.ErrParentNotFound,
		},
		{
			name: "department own parent",
			id:   "1",
			updateReq: &models.UpdateDepartmentRequest{
				ParentID: func() *int { i := 1; return &i }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				currentDept := &models.Department{ID: 1, Name: "Test"}
				m.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentOwnParent,
		},
		{
			name: "department subtree",
			id:   "1",
			updateReq: &models.UpdateDepartmentRequest{
				ParentID: func() *int { i := 5; return &i }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				currentDept := &models.Department{ID: 1, Name: "Test"}
				m.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
				m.EXPECT().DepartmentExists(5).Return(true, nil).Times(1)
				m.EXPECT().IsChildOf(5, 1).Return(true, nil).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentSubtree,
		},
		{
			name: "duplicate name in parent",
			id:   "1",
			updateReq: &models.UpdateDepartmentRequest{
				Name: func() *string { s := "Engineering"; return &s }(),
			},
			setupMock: func(m *mocks.MockOrganizationRepositoryInterface) {
				currentDept := &models.Department{ID: 1, Name: "Test", ParentID: func() *int { i := 2; return &i }()}
				m.EXPECT().GetDepartment(1, false).Return(currentDept, nil).Times(1)
				m.EXPECT().DepartmentNameExistsInParent("Engineering", gomock.Any(), 1).Return(true, nil).Times(1)
			},
			expectedErr: suberrors.ErrDepartmentNameExistsInParent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			repo := mocks.NewMockOrganizationRepositoryInterface(ctl)
			tc.setupMock(repo)

			srv := NewOrganizationService(context.Background(), repo)
			department, err := srv.PatchDepartment(tc.id, tc.updateReq)

			require.Error(t, err)
			require.Nil(t, department)
			require.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
