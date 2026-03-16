package suberrors

import "errors"

var (
	ErrDepartmentNotFound           = errors.New("department not found")
	ErrReassignDepartmentNotFound   = errors.New("reassign_department not found")
	ErrNilDepartment                = errors.New("nil department")
	ErrNilEmployee                  = errors.New("nil employee")
	ErrInvalidID                    = errors.New("invalid id")
	ErrInvalidDepth                 = errors.New("invalid depth parameter")
	ErrInvalidIncludeEmployees      = errors.New("invalid include_employees parameter")
	ErrParentNotFound               = errors.New("parent not found")
	ErrDepartmentOwnParent          = errors.New("department cannot be its own parent")
	ErrDepartmentSubtree            = errors.New("department cannot be its subtree")
	ErrDepartmentNameExistsInParent = errors.New("department with this name already exists in parent")
	ErrInvalidDepartmentName        = errors.New("invalid department name")
	ErrInvalidEmployeeFullName      = errors.New("invalid employee full name")
	ErrInvalidEmployeePosition      = errors.New("invalid employee position")
	ErrNilDepartmentUpdate          = errors.New("nil department update")
	ErrEmptyMode                    = errors.New("empty mode")
	ErrInvalidMode                  = errors.New("invalid mode")
	ErrReassignDepartmentInvalidId  = errors.New("reassign to department id is invalid")
	ErrReassignToSelf               = errors.New("cannot reassign to the same department being deleted")
	ErrReassignToChild              = errors.New("reassign to child department is invalid")
)
