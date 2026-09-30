import { execFileSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import type { AssignableRole } from "@/pages/team/helpers";

const stackName = `shellhub-e2e-${process.env.E2E_STACK_NAME || "default"}`;

export function composeExec(service: string, args: string[], input?: string) {
  return execFileSync(
    "docker",
    ["compose", "-p", stackName, "exec", "-T", service, ...args],
    {
      input,
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
      timeout: 30_000,
    },
  ).trim();
}

function serverAdmin(...args: string[]) {
  return composeExec("server", ["/server", "admin", ...args]);
}

function sql(
  query: string,
  vars: Record<string, string>,
  flags: string[] = [],
) {
  return composeExec(
    "postgres",
    [
      "sh",
      "-c",
      'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 "$@"',
      "psql",
      ...flags,
      ...Object.entries(vars).flatMap(([name, value]) => [
        "-v",
        `${name}=${value}`,
      ]),
    ],
    query,
  );
}

export const password = "e2e-password";

export const buildShortId = () => randomUUID().slice(0, 8);

export const buildRandomEmail = (prefix: string) =>
  `${prefix}-${buildShortId()}@e2e.test`;

export function createUser(prefix: string, { admin = false } = {}) {
  const id = buildShortId();
  const user = {
    username: `e2e-${prefix}-${id}`,
    email: `${prefix}-${id}@e2e.test`,
  };
  serverAdmin(
    "user",
    "create",
    user.username,
    password,
    user.email,
    ...(admin ? ["--admin"] : []),
  );
  return user;
}

export function createNamespace(owner: string, name: string, tenant: string) {
  serverAdmin("namespace", "create", name, owner, tenant);
}

export function addMember(
  username: string,
  namespace: string,
  role: AssignableRole,
) {
  serverAdmin("namespace", "member", "add", username, namespace, role);
}

export function expireInvitation(tenant: string) {
  const out = sql(
    "UPDATE membership_invitations SET expires_at = now() - interval '1 day' WHERE tenant_id = :'tenant';",
    { tenant },
  );
  if (out !== "UPDATE 1") {
    throw new Error(
      `expected to expire 1 invitation in ${tenant}, got "${out}"`,
    );
  }
}

export function readUserInvitationStatus(email: string) {
  const status = sql(
    "SELECT status FROM user_invitations WHERE email = :'email';",
    { email },
    ["-tA"],
  );
  if (!status) throw new Error(`expected a user invitation for ${email}`);
  return status;
}
