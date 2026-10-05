import {
  type ExecFileSyncOptionsWithStringEncoding,
  execFileSync,
} from "node:child_process";
import { createHash, randomUUID } from "node:crypto";
import type { NamespaceSettings } from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import { requireEnv } from "./env";

const stackName = requireEnv("E2E_COMPOSE_PROJECT");

function compose(
  args: string[],
  options: Omit<ExecFileSyncOptionsWithStringEncoding, "encoding"> = {},
) {
  return execFileSync("docker", ["compose", "-p", stackName, ...args], {
    stdio: ["pipe", "pipe", "pipe"],
    timeout: 30_000,
    ...options,
    encoding: "utf-8",
  });
}

export function composeExec(service: string, args: string[], input?: string) {
  return compose(["exec", "-T", service, ...args], { input }).trim();
}

export function composeLogs(service: string) {
  return compose(["logs", "--no-log-prefix", service], {
    maxBuffer: 64 * 1024 * 1024,
  });
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

export function buildUserIdentity(prefix: string) {
  const id = buildShortId();
  return { username: `e2e-${prefix}-${id}`, email: `${prefix}-${id}@e2e.test` };
}

export function createUserWithCli(prefix: string, { admin = false } = {}) {
  const user = buildUserIdentity(prefix);
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

export type NamespaceOptions = {
  sshAccessMode?: NamespaceSettings["ssh_access_mode"];
};

export function createNamespace(
  owner: string,
  name: string,
  tenant: string,
  { sshAccessMode }: NamespaceOptions = {},
) {
  serverAdmin(
    "namespace",
    "create",
    name,
    owner,
    tenant,
    ...(sshAccessMode ? [`--ssh-access-mode=${sshAccessMode}`] : []),
  );
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

export const mfaSecret = "JBSWY3DPEHPK3PXP";

const hashRecoveryCode = (code: string) =>
  createHash("sha256").update(code).digest("hex");

export function enableMFA(
  username: string,
  { recoveryEmail = "", recoveryCodes = [] as string[] } = {},
) {
  const out = sql(
    `UPDATE users SET mfa_enabled = true, mfa_secret = :'secret',
       security_email = NULLIF(:'recovery_email', ''),
       mfa_recovery_codes = string_to_array(NULLIF(:'codes', ''), ',')
     WHERE username = :'username';`,
    {
      username,
      secret: mfaSecret,
      recovery_email: recoveryEmail,
      codes: recoveryCodes.map(hashRecoveryCode).join(","),
    },
  );
  if (out !== "UPDATE 1") {
    throw new Error(`expected to enable MFA for ${username}, got "${out}"`);
  }
}

export function countRecoveryCodes(username: string) {
  const count = sql(
    "SELECT coalesce(cardinality(mfa_recovery_codes), 0) FROM users WHERE username = :'username';",
    { username },
    ["-tA"],
  );
  if (!count) throw new Error(`expected a user named ${username}`);
  return Number(count);
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

export function setBillingCustomer(tenant: string, customer: string) {
  const out = sql(
    "UPDATE namespaces SET billing = jsonb_build_object('customer_id', :'customer') WHERE id = :'tenant';",
    { tenant, customer },
  );
  if (out !== "UPDATE 1") {
    throw new Error(
      `expected to set the billing customer of ${tenant}, got "${out}"`,
    );
  }
}
