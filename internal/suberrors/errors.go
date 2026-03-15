package suberrors

import "errors"

var (
	ErrDepartmentNotFound           = errors.New("department not found")
	ErrReassignDepartmentNotFound   = errors.New("reassign_department not found")
	ErrNilDepartment                = errors.New("nil department")
	ErrNilEmployee                  = errors.New("nil employee")
	ErrInvalidID                    = errors.New("invalid id")
	ErrParentNotFound               = errors.New("parent not found")
	ErrDepartmentOwnParent          = errors.New("department cannot be its own parent")
	ErrDepartmentSubtree            = errors.New("department cannot be its subtree")
	ErrDepartmentNameExistsInParent = errors.New("department with this name already exists in parent")
	ErrInvalidDepartmentName        = errors.New("invalid department name")
	ErrNilDepartmentUpdate          = errors.New("nil department update")
	ErrEmptyMode                    = errors.New("empty mode")
	ErrInvalidMode                  = errors.New("invalid mode")
	ErrReassignDepartmentInvalidId  = errors.New("reassign to department id is invalid")
	ErrReassignToChild              = errors.New("reassign to child department is invalid")
)
