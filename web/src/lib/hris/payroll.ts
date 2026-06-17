import { api } from '@/lib/api';

export interface SalaryComponent {
  id: string;
  companyId: string;
  name: string;
  code: string;
  type: 'earning' | 'deduction';
  isTaxable: boolean;
  isFixed: boolean;
  defaultAmount: number;
  isActive: boolean;
  createdAt: string;
}

export interface EmployeeSalary {
  id: string;
  employeeId: string;
  componentId: string;
  componentName?: string;
  componentCode?: string;
  componentType?: string;
  amount: number;
  effectiveDate: string;
  createdAt: string;
}

export interface PayrollRun {
  id: string;
  companyId: string;
  periodStart: string;
  periodEnd: string;
  status: 'draft' | 'calculated' | 'approved' | 'paid';
  totalGross: number;
  totalDeductions: number;
  totalNet: number;
  employeeCount: number;
  notes?: string;
  runBy?: string;
  runByName?: string;
  approvedBy?: string;
  approvedAt?: string;
  createdAt: string;
}

export interface PayslipLine {
  componentName: string;
  componentCode: string;
  type: 'earning' | 'deduction';
  amount: number;
}

export interface Payslip {
  id: string;
  payrollRunId: string;
  employeeId: string;
  employeeName?: string;
  employeeCode?: string;
  grossPay: number;
  totalDeductions: number;
  netPay: number;
  breakdown: PayslipLine[];
  createdAt: string;
  periodStart?: string;
  periodEnd?: string;
}

// Salary Components
export function listSalaryComponents(): Promise<SalaryComponent[]> {
  return api<SalaryComponent[]>('/hris/salary-components');
}

export function createSalaryComponent(data: Partial<SalaryComponent>): Promise<SalaryComponent> {
  return api<SalaryComponent>('/hris/salary-components', { method: 'POST', body: JSON.stringify(data) });
}

export function updateSalaryComponent(id: string, data: Partial<SalaryComponent>): Promise<void> {
  return api(`/hris/salary-components/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteSalaryComponent(id: string): Promise<void> {
  return api(`/hris/salary-components/${id}`, { method: 'DELETE' });
}

// Employee Salary
export function getEmployeeSalary(employeeId: string): Promise<EmployeeSalary[]> {
  return api<EmployeeSalary[]>(`/hris/employees/${employeeId}/salary`);
}

export function setEmployeeSalary(
  employeeId: string,
  items: { componentId: string; amount: number; effectiveDate?: string }[],
): Promise<void> {
  return api(`/hris/employees/${employeeId}/salary`, { method: 'PUT', body: JSON.stringify(items) });
}

// Payroll Runs
export function listPayrollRuns(): Promise<PayrollRun[]> {
  return api<PayrollRun[]>('/hris/payroll-runs');
}

export function createPayrollRun(data: {
  periodStart: string;
  periodEnd: string;
  notes?: string;
}): Promise<PayrollRun> {
  return api<PayrollRun>('/hris/payroll-runs', { method: 'POST', body: JSON.stringify(data) });
}

export function calculatePayrollRun(id: string): Promise<void> {
  return api(`/hris/payroll-runs/${id}/calculate`, { method: 'POST' });
}

export function approvePayrollRun(id: string): Promise<void> {
  return api(`/hris/payroll-runs/${id}/approve`, { method: 'POST' });
}

export function markPayrollPaid(id: string): Promise<void> {
  return api(`/hris/payroll-runs/${id}/mark-paid`, { method: 'POST' });
}

// Payslips
export function listPayslips(runId: string): Promise<Payslip[]> {
  return api<Payslip[]>(`/hris/payroll-runs/${runId}/payslips`);
}

export function getMyPayslips(): Promise<Payslip[]> {
  return api<Payslip[]>('/hris/payslips/my');
}

export function getPayslip(payslipId: string): Promise<Payslip> {
  return api<Payslip>(`/hris/payslips/${payslipId}`);
}
