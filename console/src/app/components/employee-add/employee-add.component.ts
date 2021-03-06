import { Component, OnInit } from '@angular/core';

import {NgbModal, NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

import { FormGroup, FormControl, Validators } from '@angular/forms';


import { ApiService } from 'src/app/api.service';

import { ToastrService } from 'ngx-toastr';

import {Employee} from '../../api.model';

import * as config from '../../../../auth_config.json';


@Component({
  selector: 'app-employee-add',
  templateUrl: './employee-add.component.html',
  styleUrls: ['./employee-add.component.css']
})
export class EmployeeAddComponent implements OnInit {

  loading = false;

  createdEmployee: Employee;
  createdEmployeeError: any;

  employeeAddForm: FormGroup
  update: boolean = false;

  constructor(
    public activeModal: NgbActiveModal,
    private api: ApiService,
    private toastr: ToastrService
  ) { }

  ngOnInit(): void {
    this.employeeAddForm = new FormGroup({
      name: new FormControl(this.createdEmployee?.name, [
        Validators.required,
      ]),
      title: new FormControl(this.createdEmployee?.title, [
        Validators.required,
      ]),
    });
  }
  
  addEmployee(employee: Employee) {
    this.api
    .createEmployee$(employee)
    .subscribe(
      (res) => {
        this.createdEmployee = res
        this.loading = false
      },
      error => {
        this.createdEmployeeError = error.error.message || error.message
        this.loading = false
      }
    );
  }

  onSubmit() {
    this.createdEmployeeError = null
    this.loading = true
    var employee: Employee = this.employeeAddForm.value
    this.addEmployee(employee)
  }

  onUpdate() {
    this.createdEmployeeError = null
    this.loading = true
    var employee: Employee = this.employeeAddForm.value
    this.updateEmployee(employee)
  }

  totalText(id: string, name: string) {
    let total = `${id} ${name}`
    if (total.length > 22) {
      total = total.slice(0, 21) + "."
    }
    return total
  }

  updateEmployee(employee: Employee) {
    this.api
    .updateEmployee$(employee, this.createdEmployee.id.toString())
    .subscribe(
      (res) => {
        this.createdEmployee = res
        this.employeeAddForm.markAsPristine()
        this.loading = false
      },
      error => {
        this.createdEmployeeError = error.error.message || error.message
        this.loading = false
      }
    );
  }
}
