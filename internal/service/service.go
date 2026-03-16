package service

import (
	"TestHitalent2/internal/models"
	"TestHitalent2/internal/suberrors"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type OrganizationRepositoryInterface interface {
	CreateDepartment(department *models.Department) (*models.Department, error)
	CreateEmployee(employee *models.Employee) (*models.Employee, error)
	GetDepartment(id int, includeEmployees bool) (*models.Department, error)
	PatchDepartment(department *models.Department) (*models.Department, error)
	DeleteDepartmentCascade(id int) error
	DeleteDepartmentReassign(id int, intReassignToDepartmentId int) error
	GetChildrenByParentID(id int, includeEmployees bool) ([]*models.Department, error)
	DepartmentExists(id int) (bool, error)
	IsChildOf(departmentID int, parentID int) (bool, error)
	DepartmentNameExistsInParent(name string, parentID *int, excludeID int) (bool, error)
}

type OrganizationService struct {
	repo     OrganizationRepositoryInterface
	ctx      context.Context
	validate *validator.Validate
}

func NewOrganizationService(ctx context.Context, repo OrganizationRepositoryInterface) *OrganizationService {
	return &OrganizationService{
		repo:     repo,
		ctx:      ctx,
		validate: validator.New(),
	}
}

func (s *OrganizationService) CreateDepartment(department *models.Department) (*models.Department, error) {
	if department == nil {
		return nil, suberrors.ErrNilDepartment
	}

	department.Name = strings.TrimSpace(department.Name)
	if len(department.Name) < 1 || len(department.Name) > 200 {
		return nil, suberrors.ErrInvalidDepartmentName
	}

	if department.ParentID != nil {
		exists, err := s.repo.DepartmentExists(*department.ParentID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, suberrors.ErrParentNotFound
		}
	}

	exists, err := s.repo.DepartmentNameExistsInParent(department.Name, department.ParentID, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, suberrors.ErrDepartmentNameExistsInParent
	}

	return s.repo.CreateDepartment(department)
}

func (s *OrganizationService) CreateEmployee(employee *models.Employee, id string) (*models.Employee, error) {
	if employee == nil {
		return nil, suberrors.ErrNilEmployee
	}
	if id == "" {
		return nil, suberrors.ErrInvalidID
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, suberrors.ErrInvalidID
	}
	employee.DepartmentID = intID

	employee.FullName = strings.TrimSpace(employee.FullName)
	employee.Position = strings.TrimSpace(employee.Position)

	if len(employee.FullName) < 1 || len(employee.FullName) > 200 {
		return nil, suberrors.ErrInvalidEmployeeFullName
	}
	if len(employee.Position) < 1 || len(employee.Position) > 200 {
		return nil, suberrors.ErrInvalidEmployeePosition
	}

	exists, err := s.repo.DepartmentExists(employee.DepartmentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, suberrors.ErrDepartmentNotFound
	}

	return s.repo.CreateEmployee(employee)
}

func (s *OrganizationService) GetDepartment(id string, depth string, includeEmployees string) (*models.Department, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, suberrors.ErrInvalidID
	}
	intDepth := 1
	if depth != "" {
		intDepth, err = strconv.Atoi(depth)
		if err != nil {
			return nil, suberrors.ErrInvalidDepth
		}
		if intDepth > 5 {
			intDepth = 5
		}
		if intDepth < 1 {
			intDepth = 1
		}
	}
	boolIncludeEmployees := true
	if includeEmployees != "" {
		boolIncludeEmployees, err = strconv.ParseBool(strings.TrimSpace(includeEmployees))
		if err != nil {
			return nil, suberrors.ErrInvalidIncludeEmployees
		}
	}
	department, err := s.repo.GetDepartment(intID, boolIncludeEmployees)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, suberrors.ErrDepartmentNotFound
		}
		return nil, err
	}

	err = s.LoadChildren(department, intDepth, boolIncludeEmployees)
	if err != nil {
		return nil, err
	}

	return department, nil
}

func (s *OrganizationService) LoadChildren(department *models.Department, depth int, includeEmployees bool) error {
	if depth <= 0 {
		return nil
	}

	children, err := s.repo.GetChildrenByParentID(department.ID, includeEmployees)
	if err != nil {
		return err
	}

	department.Children = children

	if depth > 1 {
		for _, child := range department.Children {
			err = s.LoadChildren(child, depth-1, includeEmployees)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *OrganizationService) PatchDepartment(id string, updateDepartmentRequest *models.UpdateDepartmentRequest) (*models.Department, error) {
	if id == "" {
		return nil, suberrors.ErrInvalidID
	}
	if updateDepartmentRequest == nil {
		return nil, suberrors.ErrNilDepartmentUpdate
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, suberrors.ErrInvalidID
	}

	currentDepartment, err := s.repo.GetDepartment(intID, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, suberrors.ErrDepartmentNotFound
		}
		return nil, err
	}

	if updateDepartmentRequest.ParentID != nil {
		if *updateDepartmentRequest.ParentID == intID {
			return nil, suberrors.ErrDepartmentOwnParent
		}

		if *updateDepartmentRequest.ParentID != 0 {
			exists, err := s.repo.DepartmentExists(*updateDepartmentRequest.ParentID)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, suberrors.ErrParentNotFound
			}

			isChild, err := s.repo.IsChildOf(*updateDepartmentRequest.ParentID, intID)
			if err != nil {
				return nil, err
			}
			if isChild {
				return nil, suberrors.ErrDepartmentSubtree
			}

			currentDepartment.ParentID = updateDepartmentRequest.ParentID
		} else {
			currentDepartment.ParentID = nil
		}
	}

	if updateDepartmentRequest.Name != nil {
		trimmedName := strings.TrimSpace(*updateDepartmentRequest.Name)

		if len(trimmedName) < 1 || len(trimmedName) > 200 {
			return nil, suberrors.ErrInvalidDepartmentName
		}

		exists, err := s.repo.DepartmentNameExistsInParent(trimmedName, currentDepartment.ParentID, intID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, suberrors.ErrDepartmentNameExistsInParent
		}

		currentDepartment.Name = trimmedName
	}

	return s.repo.PatchDepartment(currentDepartment)
}

func (s *OrganizationService) DeleteDepartment(id string, mode string, reassignToDepartmentId string) error {
	if id == "" {
		return suberrors.ErrInvalidID
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return suberrors.ErrInvalidID
	}
	if mode == "" {
		return suberrors.ErrEmptyMode
	}
	if mode != "cascade" && mode != "reassign" {
		return suberrors.ErrInvalidMode
	}
	var intReassignToDepartmentId int
	if mode == "reassign" {
		if reassignToDepartmentId == "" {
			return suberrors.ErrReassignDepartmentInvalidId
		}
		intReassignToDepartmentId, err = strconv.Atoi(reassignToDepartmentId)
		if err != nil {
			return suberrors.ErrReassignDepartmentInvalidId
		}
		if intReassignToDepartmentId == intID {
			return suberrors.ErrReassignToSelf
		}
		return s.repo.DeleteDepartmentReassign(intID, intReassignToDepartmentId)
	}

	return s.repo.DeleteDepartmentCascade(intID)
}
