import { expect } from "@playwright/test";
import { apiKeyCreate, generateInvitationLink, login } from "@/client";
import { createClient } from "@/client/client";
import { requireEnv } from "./env";
import { buildShortId } from "./seed";

type Credential = { token: string } | { apiKey: string };

const client = createClient({ baseUrl: requireEnv("E2E_BASE_URL") });

client.interceptors.error.use((error, response, request) =>
  response && request && !response.ok
    ? new Error(
        `${request.method} ${request.url}: ${response.status} ${typeof error === "string" ? error : JSON.stringify(error)}`,
      )
    : error,
);

function authHeaders(auth?: Credential): Record<string, string> {
  if (!auth) return {};
  if ("token" in auth) return { Authorization: `Bearer ${auth.token}` };
  return { "X-API-Key": auth.apiKey };
}

export function buildRequestContext(auth?: Credential) {
  return { client, headers: authHeaders(auth), throwOnError: true as const };
}

type StatusContext = Omit<
  ReturnType<typeof buildRequestContext>,
  "throwOnError"
> & { throwOnError: false };

export type Endpoint = (
  opts: StatusContext,
) => Promise<{ response?: Response }>;

export async function expectStatus(
  endpoint: Endpoint,
  auth: Credential,
  expected: number,
) {
  const { response } = await endpoint({
    ...buildRequestContext(auth),
    throwOnError: false,
  });
  const via = "token" in auth ? "token" : "API key";
  expect(response?.status, `${via} on ${response?.url}`).toBe(expected);
}

export async function expectLoginStatus(
  username: string,
  attemptedPassword: string,
  status: number,
) {
  const { response } = await login({
    ...buildRequestContext(),
    throwOnError: false,
    body: { username, password: attemptedPassword },
  });
  expect(response?.status, `signing in as ${username}`).toBe(status);
}

export async function loginAs(username: string, password: string) {
  const { data } = await login({
    ...buildRequestContext(),
    body: { username, password },
  });
  return data;
}

export async function invite(token: string, tenant: string, email: string) {
  const { data } = await generateInvitationLink({
    ...buildRequestContext({ token }),
    path: { tenant },
    body: { email, role: "observer" },
  });
  if (!data.link) throw new Error(`expected an invitation link for ${email}`);
  return { email, link: data.link };
}

export async function createApiKey(token: string) {
  const name = `e2e-key-${buildShortId()}`;
  const { data } = await apiKeyCreate({
    ...buildRequestContext({ token }),
    body: { name, expires_at: -1 },
  });
  return { name, key: data.key };
}
