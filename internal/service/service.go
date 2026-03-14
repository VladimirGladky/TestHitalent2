package service

import (
	"TestHitalent2/internal/models"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

type OrganizationRepositoryInterface interface {
	CreateDepartment(department *models.Department) (*models.Department, error)
	CreateEmployee(employee *models.Employee, id int) (*models.Employee, error)
	GetDepartment(id int, includeEmployees bool) (*models.Department, error)
	PatchDepartment(id int, department *models.Department) (*models.Department, error)
	DeleteDepartmentCascade(id int) error
	DeleteDepartmentReassign(id int, intReassignToDepartmentId int) error
	GetChildrenByParentID(id int, includeEmployees bool) ([]*models.Department, error)
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
		return nil, errors.New("department is nil")
	}
	err := s.validate.Struct(department)
	if err != nil {
		return nil, err
	}
	department.Name = strings.TrimSpace(department.Name)
	return s.repo.CreateDepartment(department)
}

func (s *OrganizationService) CreateEmployee(employee *models.Employee, id string) (*models.Employee, error) {
	if employee == nil {
		return nil, errors.New("employee is nil")
	}
	if id == "" {
		return nil, errors.New("id is empty")
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = s.validate.Struct(employee)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateEmployee(employee, intID)
}

func (s *OrganizationService) GetDepartment(id string, depth string, includeEmployees string) (*models.Department, error) {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	intDepth := 1
	if depth != "" {
		intDepth, err = strconv.Atoi(depth)
		if err != nil {
			return nil, err
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
			return nil, err
		}
	}
	department, err := s.repo.GetDepartment(intID, boolIncludeEmployees)
	if err != nil {
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

func (s *OrganizationService) PatchDepartment(id string, department *models.Department) (*models.Department, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}
	if department == nil {
		return nil, errors.New("department is nil")
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return s.repo.PatchDepartment(intID, department)
}

func (s *OrganizationService) DeleteDepartment(id string, mode string, reassignToDepartmentId string) error {
	if id == "" {
		return errors.New("id is empty")
	}
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if mode == "" {
		return errors.New("mode is empty")
	}
	if mode != "cascade" && mode != "reassign" {
		return errors.New("mode is invalid")
	}
	var intReassignToDepartmentId int
	if mode == "reassign" {
		if reassignToDepartmentId != "" {
			return errors.New("reassign to department id is invalid")
		}
		intReassignToDepartmentId, err = strconv.Atoi(reassignToDepartmentId)
		if err != nil {
			return err
		}
		return s.repo.DeleteDepartmentReassign(intID, intReassignToDepartmentId)
	}
	return s.repo.DeleteDepartmentCascade(intID)
}
