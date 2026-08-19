export type Role = "registrar" | "technician" | "reviewer" | "manager";
export type StatusTone = "neutral" | "info" | "warning" | "success" | "danger";

export interface User {
  id: string;
  username: string;
  display_name: string;
  role: Role;
}
export interface Session {
  id: string;
  device: string;
  ip_address: string;
  last_seen_at: string;
  current: boolean;
}
export interface Page<T> {
  items: T[];
  next_cursor?: string;
  previous_cursor?: string;
}
export interface Attachment {
  id: string;
  filename: string;
  size: number;
  content_type: string;
}
export interface Commission {
  id: string;
  commission_no?: string;
  status: string;
  client_name: string;
  material_grade: string;
  batch_no: string;
  sample_description: string;
  test_requirements: string;
  received_at: string;
  attachments?: Attachment[];
}
export interface Sample {
  id: string;
  sample_no: string;
  parent_id?: string;
  purpose?: string;
  label: string;
  quantity: string;
  unit: string;
  status: string;
  location?: string;
  loss_reason?: string;
  children?: Sample[];
}
export interface CustodyTransfer {
  id: string;
  sample_no: string;
  kind: string;
  status: string;
  from_user: string;
  to_user: string;
  occurred_at: string;
  from_location: string;
  to_location: string;
  condition: string;
  note?: string;
}
export interface MethodField {
  key: string;
  label: string;
  type: "decimal" | "integer" | "text" | "boolean";
  unit?: string;
  required: boolean;
  min?: string;
  max?: string;
}
export interface MethodFormula {
  key: string;
  label: string;
  expression: string;
  precision: number;
  unit?: string;
}
export interface MethodVersion {
  id: string;
  method_id: string;
  name: string;
  version: string;
  status: string;
  applicable_materials: string;
  fields: MethodField[];
  formulas: MethodFormula[];
  decision_rule: string;
  created_at: string;
}
export interface TestTask {
  id: string;
  task_no: string;
  sample_no: string;
  method_name: string;
  method_version: string;
  assignee: string;
  status: string;
  due_at?: string;
  round?: number;
}
export interface Reading {
  key: string;
  value: string;
  unit?: string;
}
export interface Result {
  key: string;
  label: string;
  value: string;
  unit?: string;
  decision?: string;
}
export interface Revision {
  version: number;
  created_at: string;
  author: string;
  reason?: string;
  changes?: Record<string, { before: string; after: string }>;
}
export interface ReviewCase {
  id: string;
  task_no: string;
  sample_no: string;
  executor: string;
  status: string;
  readings: Reading[];
  results: Result[];
  revisions: Revision[];
  attachments?: Attachment[];
}
export interface Certificate {
  id: string;
  certificate_no: string;
  sample_no: string;
  version: number;
  status: string;
  issued_at: string;
  digest: string;
  verification_code?: string;
  method_versions?: string[];
  results?: Result[];
}
export interface ApiProblem {
  code?: string;
  message: string;
  request_id?: string;
  fields?: Record<string, string>;
}
