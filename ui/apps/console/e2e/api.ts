import type { APIRequestContext } from "@playwright/test";
import type {
  LoginResponse,
  GenerateInvitationLinkResponse,
  NamespaceMemberRole,
} from "@/client";
import { requireEnv } from "./env";

const baseURL = requireEnv("E2E_BASE_URL");

interface ApiOptions {
  token?: string;
  data?: unknown;
}

async function api<T>(
  request: APIRequestContext,
  method: "get" | "post" | "patch" | "put" | "delete",
  path: string,
  opts: ApiOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {};
  if (opts.token) headers["Authorization"] = `Bearer ${opts.token}`;

  const res = await request[method](`${baseURL}${path}`, {
    headers,
    data: opts.data,
  });

  if (!res.ok()) {
    throw new Error(
      `${method.toUpperCase()} ${path}: ${res.status()} ${await res.text()}`,
    );
  }

  return res.json() as Promise<T>;
}

export function login(
  request: APIRequestContext,
  username: string,
  password: string,
) {
  return api<LoginResponse>(request, "post", "/api/login", {
    data: { username, password },
  });
}

export function createInvitationLink(
  request: APIRequestContext,
  token: string,
  tenant: string,
  email: string,
  role: Exclude<NamespaceMemberRole, "owner"> = "observer",
) {
  return api<GenerateInvitationLinkResponse>(
    request,
    "post",
    `/api/namespaces/${tenant}/invitations/links`,
    { token, data: { email, role } },
  );
}
