import { client } from "@/client/client.gen";

/**
 * Hand-written calls for three API routes the OpenAPI spec does not describe, so
 * `@hey-api/openapi-ts` emits no operation for them: `billing/portal` and `user/saml/reauth` are
 * cloud-only, and `ws/ssh/session` is registered outside the spec-walked router.
 *
 * They go through the generated `client` instance rather than a second HTTP client, so they share
 * the one interceptor pipeline every generated operation uses — bearer-token injection, expiry and
 * 401 logout, and connectivity tracking. Replace each with its generated call once the route
 * reaches a spec.
 *
 * Declare each *response* shape as an `interface`, never a `type`: the client unwraps a response
 * type that satisfies `Record<string, unknown>` down to its member (it expects the generated
 * status-code map), and only an interface escapes that unwrapping. A *query* shape needs the
 * opposite — the client types `query` as `Record<string, unknown>`, which only a `type` satisfies.
 */

interface RedirectUrl {
  url: string;
}

export function getBillingPortal(): Promise<RedirectUrl> {
  return client
    .post<RedirectUrl, unknown, true>({
      url: "/api/billing/portal",
      throwOnError: true,
    })
    .then(({ data }) => data);
}

export type SamlReauthQuery = {
  fingerprint: string;
  approval_code: string;
};

export function getSamlReauthUrl(query: SamlReauthQuery): Promise<RedirectUrl> {
  return client
    .get<RedirectUrl, unknown, true>({
      url: "/api/user/saml/reauth",
      query,
      throwOnError: true,
    })
    .then(({ data }) => data);
}

export interface WebSshSessionRequest {
  device: string;
  username: string;
  password?: string;
  fingerprint?: string;
  public_key?: string;
}

interface WebSshSession {
  token: string;
}

export function createWebSshSession(
  body: WebSshSessionRequest,
): Promise<WebSshSession> {
  return client
    .post<WebSshSession, unknown, true>({
      url: "/ws/ssh/session",
      body,
      parseAs: "json",
      throwOnError: true,
    })
    .then(({ data }) => data);
}
