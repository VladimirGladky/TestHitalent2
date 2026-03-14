package transport

import (
	"TestHitalent2/internal/config"
	"TestHitalent2/internal/models"
	"TestHitalent2/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OrganizationServiceInterface interface {
	CreateDepartment(department *models.Department) (*models.Department, error)
	CreateEmployee(employee *models.Employee, id string) (*models.Employee, error)
	GetDepartment(id string, depth string, includeEmployees string) (*models.Department, error)
	PatchDepartment(id string, department *models.Department) (*models.Department, error)
	DeleteDepartment(id string, mode string, reassignToDepartmentId string) error
}

type OrganizationServer struct {
	cfg     *config.Config
	service OrganizationServiceInterface
	ctx     context.Context
}

func NewOrganizationServer(cfg *config.Config, service OrganizationServiceInterface, ctx context.Context) *OrganizationServer {
	return &OrganizationServer{
		cfg:     cfg,
		service: service,
		ctx:     ctx,
	}
}

func (s *OrganizationServer) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/departments", CreateDepartmentsHandler(s))
	mux.HandleFunc("POST /api/v1/departments/{id}/employees", CreateEmployeesHandler(s))
	mux.HandleFunc("GET /api/v1/departments/{id}", GetDepartmentsHandler(s))
	mux.HandleFunc("PATCH /api/v1/departments/{id}", PatchDepartmentsHandler(s))
	mux.HandleFunc("DELETE /api/v1/departments/{id}", DeleteDepartmentsHandler(s))
	logger.GetLoggerFromCtx(s.ctx).Info("HTTP server is running")
	addr := s.cfg.Host + ":" + s.cfg.Port
	return http.ListenAndServe(addr, mux)
}

func CreateDepartmentsHandler(s *OrganizationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal server error 1", "description": "` + fmt.Sprint(rec) + `"}`))
				return
			}
		}()
		defer r.Body.Close()
		req := new(models.Department)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Invalid request body", "description": "` + err.Error() + `"}`))
			return
		}
		department, err := s.service.CreateDepartment(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 2", "description": "` + err.Error() + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(models.Department{ID: department.ID, Name: department.Name, CreatedAt: department.CreatedAt})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 3", "description": "` + err.Error() + `"}`))
			return
		}
	}
}

func CreateEmployeesHandler(s *OrganizationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal server error 1", "description": "` + fmt.Sprint(rec) + `"}`))
				return
			}
		}()
		id := r.PathValue("id")
		defer r.Body.Close()
		req := new(models.Employee)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Invalid request body", "description": "` + err.Error() + `"}`))
			return
		}
		employee, err := s.service.CreateEmployee(req, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 2", "description": "` + err.Error() + `"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(models.Employee{ID: employee.ID, FullName: employee.FullName, CreatedAt: employee.CreatedAt})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 3", "description": "` + err.Error() + `"}`))
			return
		}
	}
}

func GetDepartmentsHandler(s *OrganizationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal server error 1", "description": "` + fmt.Sprint(rec) + `"}`))
				return
			}
		}()
		depth := r.URL.Query().Get("depth")
		includeEmployees := r.URL.Query().Get("include_employees")
		id := r.PathValue("id")
		defer r.Body.Close()
		department, err := s.service.GetDepartment(id, depth, includeEmployees)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 2", "description": "` + err.Error() + `"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(department)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 3", "description": "` + err.Error() + `"}`))
			return
		}
	}
}

func PatchDepartmentsHandler(s *OrganizationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal server error 1", "description": "` + fmt.Sprint(rec) + `"}`))
				return
			}
		}()
		id := r.PathValue("id")
		defer r.Body.Close()
		req := new(models.Department)
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "Invalid request body", "description": "` + err.Error() + `"}`))
			return
		}
		department, err := s.service.PatchDepartment(id, req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 2", "description": "` + err.Error() + `"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(department)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 3", "description": "` + err.Error() + `"}`))
			return
		}
	}
}

func DeleteDepartmentsHandler(s *OrganizationServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error": "Internal server error 1", "description": "` + fmt.Sprint(rec) + `"}`))
				return
			}
		}()
		mode := r.URL.Query().Get("mode")
		reassignToDepartmentId := r.URL.Query().Get("reassign_to_department_id")
		id := r.PathValue("id")
		defer r.Body.Close()
		err := s.service.DeleteDepartment(id, mode, reassignToDepartmentId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error": "Internal server error 2", "description": "` + err.Error() + `"}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
