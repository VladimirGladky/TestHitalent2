package repository

import (
	"TestHitalent2/internal/models"
	"TestHitalent2/internal/suberrors"
	"context"
	"errors"

	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewOrganizationRepository(db *gorm.DB, ctx context.Context) *OrganizationRepository {
	return &OrganizationRepository{
		db:  db,
		ctx: ctx,
	}
}

func (o *OrganizationRepository) CreateDepartment(department *models.Department) (*models.Department, error) {
	if err := o.db.WithContext(o.ctx).Create(department).Error; err != nil {
		return nil, err
	}
	return department, nil
}

func (o *OrganizationRepository) CreateEmployee(employee *models.Employee) (*models.Employee, error) {
	if err := o.db.WithContext(o.ctx).Create(employee).Error; err != nil {
		return nil, err
	}
	return employee, nil
}

func (o *OrganizationRepository) DepartmentExists(id int) (bool, error) {
	var count int64
	err := o.db.WithContext(o.ctx).Model(&models.Department{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (o *OrganizationRepository) GetDepartment(id int, includeEmployees bool) (*models.Department, error) {
	var department models.Department

	query := o.db.WithContext(o.ctx)

	if includeEmployees {
		query = query.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, full_name ASC")
		})
	}

	if err := query.First(&department, id).Error; err != nil {
		return nil, err
	}

	return &department, nil
}

func (o *OrganizationRepository) GetChildrenByParentID(id int, includeEmployees bool) ([]*models.Department, error) {
	var departments []*models.Department

	query := o.db.WithContext(o.ctx).Where("parent_id = ?", id)

	if includeEmployees {
		query = query.Preload("Employees", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, full_name ASC")
		})
	}

	err := query.Find(&departments).Error
	return departments, err
}

func (o *OrganizationRepository) PatchDepartment(department *models.Department) (*models.Department, error) {
	if err := o.db.WithContext(o.ctx).Save(department).Error; err != nil {
		return nil, err
	}
	return department, nil
}

func (o *OrganizationRepository) IsChildOf(departmentID int, parentID int) (bool, error) {
	currentID := departmentID

	for {
		var dept models.Department
		if err := o.db.WithContext(o.ctx).First(&dept, currentID).Error; err != nil {
			return false, err
		}

		if dept.ParentID == nil {
			return false, nil
		}

		if *dept.ParentID == parentID {
			return true, nil
		}

		currentID = *dept.ParentID
	}
}

func (o *OrganizationRepository) DepartmentNameExistsInParent(name string, parentID *int, excludeID int) (bool, error) {
	var count int64
	query := o.db.WithContext(o.ctx).Model(&models.Department{}).
		Where("name = ?", name).
		Where("id != ?", excludeID)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

func (o *OrganizationRepository) DeleteDepartmentCascade(id int) error {
	result := o.db.
		WithContext(o.ctx).
		Delete(&models.Department{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return suberrors.ErrDepartmentNotFound
	}

	return nil
}

func (o *OrganizationRepository) DeleteDepartmentReassign(id int, reassignToID int) error {
	return o.db.WithContext(o.ctx).Transaction(func(tx *gorm.DB) error {
		var dept models.Department
		if err := tx.First(&dept, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return suberrors.ErrDepartmentNotFound
			}
			return err
		}

		var reassignDept models.Department
		if err := tx.First(&reassignDept, reassignToID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return suberrors.ErrReassignDepartmentNotFound
			}
			return err
		}

		isChild, err := o.IsChildOf(reassignToID, id)
		if err != nil {
			return err
		}
		if isChild {
			return suberrors.ErrReassignToChild
		}

		if err = tx.Model(&models.Employee{}).
			Where("department_id = ?", id).
			Update("department_id", reassignToID).Error; err != nil {
			return err
		}

		if err = tx.Model(&models.Department{}).
			Where("parent_id = ?", id).
			Update("parent_id", reassignToID).Error; err != nil {
			return err
		}

		if err = tx.Delete(&dept).Error; err != nil {
			return err
		}

		return nil
	})
}
