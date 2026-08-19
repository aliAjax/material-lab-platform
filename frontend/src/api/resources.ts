import { api, queryString, writeOptions } from "./client";
import type {
  Certificate,
  Commission,
  CustodyTransfer,
  MethodVersion,
  Page,
  ReviewCase,
  Sample,
  Session,
  TestTask,
  User,
} from "../types/domain";

type BackendUser = { id: string; username: string; displayName: string; role: "registrar" | "tester" | "reviewer" | "manager" };
type BackendSession = { id: string; userAgent: string; createdAt: string; expiresAt: string };
type BackendCommission = { id: string; reference?: string; status: string; client: { organization: string }; materialGrade: string; batchNumber: string; sampleDescription: string; requirements: string; receivedAt: string };
type BackendSample = { id: string; number: string; description: string; totalQuantity: string; unit: string; status: string };
type BackendSubsample = { id: string; label: string; purpose: string; quantity: string; unit: string; status: string; locationId?: string };
type BackendMethod = { id: string; code: string; name: string; version: number; status: string; materialScope: string; fields: { name: string; label: string; type: "number" | "text" | "boolean"; unit: string; required: boolean; min?: string; max?: string }[]; formula: string; precision: number; rule: string; createdAt: string };
type BackendTask = { id: string; subsampleId: string; methodId: string; methodVersion: number; assigneeId: string; status: string };

const mapUser = (value: BackendUser): User => ({...value, display_name: value.displayName, role: value.role === "tester" ? "technician" : value.role});
const mapSession = (value: BackendSession): Session => ({id:value.id,device:value.userAgent || "当前浏览器",ip_address:"-",last_seen_at:value.createdAt,current:false});
const mapCommission = (value: BackendCommission): Commission => ({id:value.id,commission_no:value.reference,status:value.status,client_name:value.client.organization,material_grade:value.materialGrade,batch_no:value.batchNumber,sample_description:value.sampleDescription,test_requirements:value.requirements,received_at:value.receivedAt});
const mapSample = (value: BackendSample): Sample => ({id:value.id,sample_no:value.number,label:value.description,quantity:value.totalQuantity,unit:value.unit,status:value.status});
const mapSubsample = (value: BackendSubsample): Sample => ({id:value.id,sample_no:value.id,label:value.label,purpose:value.purpose,quantity:value.quantity,unit:value.unit,status:value.status,location:value.locationId});
const mapMethod = (value: BackendMethod): MethodVersion => ({id:value.id,method_id:value.code,name:value.name,version:String(value.version),status:value.status,applicable_materials:value.materialScope,fields:value.fields.map(field=>({key:field.name,label:field.label,type:field.type === "number" ? "decimal" : field.type,unit:field.unit,required:field.required,min:field.min,max:field.max})),formulas:[{key:"result",label:"计算结果",expression:value.formula,precision:value.precision}],decision_rule:value.rule,created_at:value.createdAt});
const mapTask = (value: BackendTask): TestTask => ({id:value.id,task_no:value.id.slice(0,8),sample_no:value.subsampleId,method_name:value.methodId,method_version:String(value.methodVersion),assignee:value.assigneeId,status:value.status});
const commissionPayload = (value: Partial<Commission>) => ({organization:value.client_name || "",materialGrade:value.material_grade || "",batchNumber:value.batch_no || "",sampleDescription:value.sample_description || "",requirements:value.test_requirements || "",receivedAt:value.received_at ? new Date(value.received_at).toISOString() : new Date().toISOString()});
const splitPayload = (value: unknown) => { const raw=value as Record<string, string>; return {parts:[{label:raw.label,purpose:raw.purpose,quantity:raw.quantity,unit:raw.unit,locationId:raw.location || "bench-a"}],loss:"0",lossReason:raw.loss_reason || ""}; };
const methodPayload = (value: Partial<MethodVersion>) => ({code:value.method_id || `METHOD-${Date.now()}`,name:value.name || "未命名方法",materialScope:value.applicable_materials || "通用材料",version:Number(value.version || 1),fields:(value.fields || []).map(field=>({name:field.key,label:field.label,type:field.type === "decimal" || field.type === "integer" ? "number" : field.type,unit:field.unit || "",required:field.required,min:field.min,max:field.max})),formula:value.formulas?.[0]?.expression || "0",precision:value.formulas?.[0]?.precision || 2,rule:value.decision_rule || "result >= 0"});

export const authApi = {
  login: async (username: string, password: string) => {
    const result = await api<{ accessToken: string; refreshToken: string; user: BackendUser }>(
      "/auth/login",
      writeOptions("POST", { username, password }),
    );
    sessionStorage.setItem("lab_refresh_token", result.refreshToken);
    return { access_token: result.accessToken, user: mapUser(result.user) };
  },
  me: async () => mapUser((await api<{user: BackendUser}>("/me")).user),
  logout: () => api<void>("/auth/logout", writeOptions("POST", {sessionId: sessionStorage.getItem("lab_refresh_token")?.split(".")[0] || ""})),
  sessions: async () => (await api<BackendSession[]>("/sessions")).map(mapSession),
  revokeSession: (id: string) =>
    api<void>(`/sessions/${id}`, writeOptions("DELETE")),
};
export const commissionsApi = {
  list: (params: Record<string, string | undefined>) =>
    api<{items: BackendCommission[]; nextCursor?: string}>(`/commissions${queryString(params)}`).then(x => ({items:x.items.map(mapCommission), next_cursor:x.nextCursor})),
  get: (id: string) => api<BackendCommission>(`/commissions/${id}`).then(mapCommission),
  save: async (value: Partial<Commission>) => mapCommission(await api<BackendCommission>(value.id ? `/commissions/${value.id}` : "/commissions", writeOptions(value.id ? "PATCH" : "POST", commissionPayload(value)))),
  submit: async (id: string) => {
    const result = await api<{commission: BackendCommission; sample: BackendSample}>(`/commissions/${id}/submit`, writeOptions("POST", {totalQuantity:"10",unit:"piece"}));
    return {...mapCommission(result.commission), id: result.sample.id};
  },
};
export const samplesApi = {
  list: (params: Record<string, string | undefined>) =>
    api<{items: BackendSample[]; nextCursor?: string}>(`/samples${queryString(params)}`).then(x=>({items:x.items.map(mapSample),next_cursor:x.nextCursor,previous_cursor:undefined})),
  get: async (id: string) => { const x=await api<{sample:BackendSample;subsamples:BackendSubsample[]}>(`/samples/${id}`); return {...mapSample(x.sample),children:x.subsamples.map(mapSubsample)} },
  split: (id: string, value: unknown) =>
    api<BackendSubsample[]>(`/samples/${id}/splits`, writeOptions("POST", splitPayload(value))).then(values => mapSubsample(values[0]!)),
};
export const custodyApi = {
  list: (params: Record<string, string | undefined>) =>
    api<Page<CustodyTransfer>>(`/custody-transfers${queryString(params)}`),
  create: (value: unknown) =>
    api<CustodyTransfer>("/custody-transfers", writeOptions("POST", value)),
  confirm: (id: string) =>
    api<CustodyTransfer>(
      `/custody-transfers/${id}/confirm`,
      writeOptions("POST"),
    ),
  reverse: (id: string, reason: string) =>
    api<CustodyTransfer>(
      `/custody-transfers/${id}/reversals`,
      writeOptions("POST", { reason }),
    ),
};
export const methodsApi = {
  list: () => api<{items:BackendMethod[]}>("/methods").then(x=>({items:x.items.map(mapMethod)})),
  get: (id: string) => api<BackendMethod>(`/methods/${id}`).then(mapMethod),
  save: (value: Partial<MethodVersion>) =>
    value.id
      ? api<MethodVersion>(
          `/methods/${value.id}`,
          writeOptions("PUT", value),
        )
      : api<BackendMethod>("/methods", writeOptions("POST", methodPayload(value))).then(mapMethod),
  validate: (id: string) =>
    api<{ valid: boolean; errors?: string[] }>(
      `/methods/${id}/validate`,
      writeOptions("POST"),
    ),
  publish: (id: string) =>
    api<BackendMethod>(`/methods/${id}/publish`, writeOptions("POST")).then(mapMethod),
};
export const tasksApi = {
  list: (params: Record<string, string | undefined>) =>
    api<{items:BackendTask[]}>(`/tasks${queryString(params)}`).then(x=>({items:x.items.map(mapTask),next_cursor:undefined,previous_cursor:undefined})),
  get: (id: string) =>
    api<
      TestTask & { method: MethodVersion; readings?: Record<string, string> }
    >(`/tasks/${id}`),
  start: (id: string) =>
    api<TestTask>(`/tasks/${id}/start`, writeOptions("POST")),
  saveReadings: (id: string, readings: Record<string, string>) =>
    api(`/tasks/${id}/rounds`, writeOptions("POST", { readings })),
  preview: (id: string, readings: Record<string, string>) =>
    api<{ results: Record<string, string>; warnings?: string[] }>(
      `/tasks/${id}/rounds`,
      writeOptions("POST", { readings }),
    ),
  submit: (id: string) =>
    api<TestTask>(`/tasks/${id}/submit-review`, writeOptions("POST")),
  retest: (id: string, reason: string, issue_type: string) =>
    api(
      `/tasks/${id}/retests`,
      writeOptions("POST", { reason, issue_type }),
    ),
};
export const reviewsApi = {
  list: (params: Record<string, string | undefined>) =>
    api<Page<ReviewCase>>(`/reviews${queryString(params)}`),
  get: (id: string) => api<ReviewCase>(`/reviews/${id}`),
  decide: (id: string, decision: "approve" | "return", comment: string) =>
    api(`/reviews/${id}/decision`, writeOptions("POST", { decision, comment })),
  decideRetest: (id: string, decision: "approve" | "reject", comment: string) =>
    api(`/retests/${id}/decision`, writeOptions("POST", { decision, comment })),
};
export const certificatesApi = {
  list: () => api<Page<Certificate>>("/certificates"),
  get: (id: string) => api<Certificate>(`/certificates/${id}`),
  issue: (commission_id: string) =>
    api<Certificate>("/certificates/issue", writeOptions("POST", { requestId: commission_id })),
  void: (id: string, reason: string) =>
    api(`/certificates/${id}/void`, writeOptions("POST", { reason })),
  verify: (code: string) =>
    api<Certificate>(`/public/certificates/verify${queryString({ code })}`),
};
