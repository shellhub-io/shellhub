import { generateInvitationLink, login } from "@/client";
import { createClient } from "@/client/client";
import { requireEnv } from "./env";

type Credential = { token: string };

const client = createClient({ baseUrl: requireEnv("E2E_BASE_URL") });

client.interceptors.error.use((error, response, request) =>
  response && request && !response.ok
    ? new Error(
        `${request.method} ${request.url}: ${response.status} ${typeof error === "string" ? error : JSON.stringify(error)}`,
      )
    : error,
);

function authHeaders(auth?: Credential): Record<string, string> {
  return auth ? { Authorization: `Bearer ${auth.token}` } : {};
}

export function buildRequestContext(auth?: Credential) {
  return { client, headers: authHeaders(auth), throwOnError: true as const };
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
