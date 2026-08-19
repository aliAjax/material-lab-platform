import type { ApiProblem } from "../types/domain";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api/v1";

export class ApiError extends Error {
  status: number;
  problem: ApiProblem;
  constructor(status: number, problem: ApiProblem) {
    super(problem.message || `请求失败 (${status})`);
    this.status = status;
    this.problem = problem;
  }
}

let accessToken = sessionStorage.getItem("lab_access_token") || "";
let refreshPromise: Promise<boolean> | undefined;
export const authToken = {
  get: () => accessToken,
  set: (token: string) => {
    accessToken = token;
    token
      ? sessionStorage.setItem("lab_access_token", token)
      : sessionStorage.removeItem("lab_access_token");
  },
};

async function refreshAccess(): Promise<boolean> {
  if (!refreshPromise)
    refreshPromise = fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    })
      .then(async (response) => {
        if (!response.ok) return false;
        const body = await response.json();
        authToken.set(body.access_token);
        return true;
      })
      .catch(() => false)
      .finally(() => {
        refreshPromise = undefined;
      });
  return refreshPromise;
}

export async function api<T>(
  path: string,
  init: RequestInit & { retryAuth?: boolean } = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  if (!(init.body instanceof FormData))
    headers.set("Content-Type", "application/json");
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
  headers.set("X-Request-ID", crypto.randomUUID());
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });
  if (
    response.status === 401 &&
    init.retryAuth !== false &&
    (await refreshAccess())
  )
    return api<T>(path, { ...init, retryAuth: false });
  if (!response.ok) {
    const body = await response
      .json()
      .catch(() => ({ message: response.statusText || "服务暂时不可用" }));
    const problem = body.error || body;
    if (response.status === 401)
      window.dispatchEvent(new Event("session-expired"));
    throw new ApiError(response.status, problem);
  }
  if (response.status === 204) return undefined as T;
  const body = await response.json();
  return (Object.prototype.hasOwnProperty.call(body, "data") ? body.data : body) as T;
}

export function writeOptions(
  method: "POST" | "PUT" | "PATCH" | "DELETE",
  body?: unknown,
): RequestInit {
  return {
    method,
    body: body === undefined ? undefined : JSON.stringify(body),
    headers: { "Idempotency-Key": crypto.randomUUID() },
  };
}

export function queryString(values: Record<string, string | undefined>) {
  const params = new URLSearchParams();
  Object.entries(values).forEach(
    ([key, value]) => value && params.set(key, value),
  );
  const query = params.toString();
  return query ? `?${query}` : "";
}
