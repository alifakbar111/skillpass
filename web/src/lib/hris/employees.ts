import { api } from '@/lib/api';

export interface Employee {
  id: string;
  companyId: string;
  userId?: string;
  employeeIdNumber: string;
  firstName: string;
  lastName: string;
  email: string;
  phone?: string;
  dateOfBirth?: string;
  gender?: string;
  maritalStatus?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  nationalId?: string;
  npwp?: string;
  bpjsKesehatanId?: string;
  bpjsKetenagakerjaanId?: string;
  bankName?: string;
  bankAccountNumber?: string;
  bankAccountHolder?: string;
  emergencyContactName?: string;
  emergencyContactPhone?: string;
  emergencyContactRelation?: string;
  employmentType: string;
  employmentStatus: string;
  joinDate: string;
  endDate?: string;
  departmentId?: string;
  positionId?: string;
  branchId?: string;
  managerId?: string;
  baseSalary?: number;
  departmentName?: string;
  positionName?: string;
  branchName?: string;
  createdAt: string;
  updatedAt: string;
}

export interface EmployeeListResult {
  employees: Employee[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateEmployeeRequest {
  firstName: string;
  lastName?: string;
  email: string;
  phone?: string;
  employmentType: string;
  joinDate: string;
  departmentId?: string;
  positionId?: string;
  branchId?: string;
  managerId?: string;
  baseSalary?: number;
}

export interface UpdateEmployeeRequest {
  firstName?: string;
  lastName?: string;
  email?: string;
  phone?: string;
  dateOfBirth?: string;
  gender?: string;
  maritalStatus?: string;
  address?: string;
  city?: string;
  province?: string;
  postalCode?: string;
  nationalId?: string;
  npwp?: string;
  bpjsKesehatanId?: string;
  bpjsKetenagakerjaanId?: string;
  bankName?: string;
  bankAccountNumber?: string;
  bankAccountHolder?: string;
  emergencyContactName?: string;
  emergencyContactPhone?: string;
  emergencyContactRelation?: string;
  employmentType?: string;
  employmentStatus?: string;
  endDate?: string;
  departmentId?: string;
  positionId?: string;
  branchId?: string;
  managerId?: string;
  baseSalary?: number;
}

interface ListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  departmentId?: string;
  branchId?: string;
  search?: string;
}

export function listEmployees(params: ListParams = {}): Promise<EmployeeListResult> {
  const query = new URLSearchParams();
  if (params.page) query.set('page', String(params.page));
  if (params.pageSize) query.set('pageSize', String(params.pageSize));
  if (params.status) query.set('status', params.status);
  if (params.departmentId) query.set('departmentId', params.departmentId);
  if (params.branchId) query.set('branchId', params.branchId);
  if (params.search) query.set('search', params.search);
  const qs = query.toString();
  return api<EmployeeListResult>(`/hris/employees${qs ? `?${qs}` : ''}`);
}

export function getEmployee(id: string): Promise<Employee> {
  return api<Employee>(`/hris/employees/${id}`);
}

export function createEmployee(data: CreateEmployeeRequest): Promise<Employee> {
  return api<Employee>('/hris/employees', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateEmployee(id: string, data: UpdateEmployeeRequest): Promise<Employee> {
  return api<Employee>(`/hris/employees/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}
