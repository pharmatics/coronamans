package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Employee struct {
	ID       uint64 `json:"id,omitempty"`
	Name     string `json:"name,omitempty" gorm:"not null;uniqueIndex;size:191"`
	Title    string `json:"title,omitempty" gorm:"not null;index"`
	LoggedIn bool   `json:"loggedIn" gorm:"not null"`

	CreatedAt time.Time `json:"createdAt,omitempty" gorm:"index"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

func (e Employee) validate() *Status {
	if e.Name == "" {
		return &Status{
			Status:  StatusFailure,
			Message: "Missing employee name",
			Code:    http.StatusBadRequest,
		}
	}

	if e.Title == "" {
		return &Status{
			Status:  StatusFailure,
			Message: "Missing employee title",
			Code:    http.StatusBadRequest,
		}
	}
	return nil
}

func getNewEmployeeID() uint64 {
	var employeeID uint64
	row := db.Raw("SELECT NEXTVAL(employee_serial)").Row()
	row.Scan(&employeeID)
	return employeeID
}

func CreateEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var e Employee

	err := json.NewDecoder(r.Body).Decode(&e)
	if err != nil {
		status := Status{
			Status:  StatusFailure,
			Message: "Invalid employee request",
			Code:    http.StatusBadRequest,
			Details: e,
		}
		ResponseJSON(status, w, http.StatusBadRequest)
		return
	}

	if status := e.validate(); status != nil {
		ResponseJSON(status, w, http.StatusBadRequest)
		return
	}

	e.ID = getNewEmployeeID()
	e.LoggedIn = false

	if err := db.Create(&e).Error; err != nil {
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				status := Status{
					Status:  StatusFailure,
					Message: fmt.Sprintf("Employee with name '%s' already exists", e.Name),
					Code:    http.StatusInternalServerError,
					Details: err,
				}
				ResponseJSON(status, w, http.StatusInternalServerError)
				return
			}
		}
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't create employee: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}
	ResponseJSON(e, w, 200)
}

func UpdateEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	barcode := ps.ByName("barcode")
	var employee Employee

	if err := db.First(&employee, "id = ?", barcode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			status := Status{
				Status:  StatusFailure,
				Message: "Employee not found!",
				Code:    http.StatusNotFound,
			}
			ResponseJSON(status, w, http.StatusNotFound)
			return
		} else {
			status := Status{
				Status:  StatusFailure,
				Message: fmt.Sprintf("Couldn't fetch employee: %v", err.Error()),
				Code:    http.StatusInternalServerError,
				Details: err,
			}
			ResponseJSON(status, w, http.StatusInternalServerError)
			return
		}
	}

	var updateEmployee Employee

	err := json.NewDecoder(r.Body).Decode(&updateEmployee)
	if err != nil {
		status := Status{
			Status:  StatusFailure,
			Message: "Invalid employee request",
			Code:    http.StatusBadRequest,
			Details: updateEmployee,
		}
		ResponseJSON(status, w, http.StatusBadRequest)
		return
	}

	// update struct
	employee.Name = updateEmployee.Name
	employee.Title = updateEmployee.Title

	if status := employee.validate(); status != nil {
		ResponseJSON(status, w, http.StatusBadRequest)
		return
	}

	tx := db.Begin()

	if err := tx.Save(&employee).Error; err != nil {
		tx.Rollback()
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				status := Status{
					Status:  StatusFailure,
					Message: fmt.Sprintf("Employee with name '%s' already exists", updateEmployee.Name),
					Code:    http.StatusInternalServerError,
					Details: err,
				}
				ResponseJSON(status, w, http.StatusInternalServerError)
				return
			}
		}
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't update employee: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}

	if err := tx.Model(&Attendance{}).Where("barcode = ?", employee.ID).Updates(Attendance{Name: employee.Name, Title: employee.Title}).Error; err != nil {
		tx.Rollback()
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't update employee attendance: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't update employee and his attendance: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}

	ResponseJSON(employee, w, 200)
}

func GetEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	barcode := ps.ByName("barcode")
	var e Employee

	if err := db.First(&e, "id = ?", barcode).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			status := Status{
				Status:  StatusFailure,
				Message: "Employee not found!",
				Code:    http.StatusNotFound,
			}
			ResponseJSON(status, w, http.StatusNotFound)
			return
		} else {
			status := Status{
				Status:  StatusFailure,
				Message: fmt.Sprintf("Couldn't fetch employee: %v", err.Error()),
				Code:    http.StatusInternalServerError,
				Details: err,
			}
			ResponseJSON(status, w, http.StatusInternalServerError)
			return
		}
	}
	ResponseJSON(e, w, 200)
}

func GetEmployees(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var employees []Employee
	if err := db.Find(&employees).Error; err != nil {
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't fetch employees: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}

	ResponseJSON(employees, w, 200)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	barcode := ps.ByName("barcode")
	var e Employee

	if err := db.Delete(&e, barcode).Error; err != nil {
		status := Status{
			Status:  StatusFailure,
			Message: fmt.Sprintf("Couldn't delete employee: %v", err.Error()),
			Code:    http.StatusInternalServerError,
			Details: err,
		}
		ResponseJSON(status, w, http.StatusInternalServerError)
		return
	}

	status := Status{
		Status:  StatusSuccess,
		Message: "Employee Deleted",
		Code:    200,
		Details: barcode,
	}
	ResponseJSON(status, w, 200)
}
