import {
  type ExecFileSyncOptionsWithStringEncoding,
  execFileSync,
} from "node:child_process";
import { randomUUID } from "node:crypto";
import type { NamespaceSettings } from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import { requireEnv } from "./env";

const stackName = requireEnv("E2E_COMPOSE_PROJECT");
const agentImage = requireEnv("E2E_AGENT_IMAGE");

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

export function runInGateway(args: string[], command: string[] = []) {
  const gateway = compose(["ps", "-q", "gateway"]).trim();
  if (!gateway) throw new Error(`expected a running gateway in ${stackName}`);
  return execFileSync(
    "docker",
    [
      "run",
      "--detach",
      `--label=com.docker.compose.project=${stackName}`,
      "--label=com.docker.compose.service=agent",
      `--label=io.shellhub.e2e.run=${stackName}`,
      `--network=container:${gateway}`,
      ...args,
      agentImage,
      ...command,
    ],
    { encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"], timeout: 30_000 },
  ).trim();
}

export function startAgent(tenant: string, hostname: string) {
  return runInGateway([
    "--env=SHELLHUB_SERVER_ADDRESS=http://localhost",
    `--env=SHELLHUB_TENANT_ID=${tenant}`,
    `--env=SHELLHUB_PREFERRED_HOSTNAME=${hostname}`,
    `--env=SHELLHUB_PREFERRED_IDENTITY=${hostname}`,
    "--env=SHELLHUB_PRIVATE_KEY=/tmp/shellhub.key",
    "--env=SHELLHUB_KEEPALIVE_INTERVAL=1",
  ]);
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

export function deleteLicenses() {
  const out = sql("DELETE FROM licenses;", {});
  if (!out.startsWith("DELETE ") || out === "DELETE 0") {
    throw new Error(`expected to delete the installed license, got "${out}"`);
  }
}

function evictCachedSystem() {
  composeExec("redis", ["valkey-cli", "DEL", "system"]);
}

export function reopenSetup() {
  const out = sql(
    "UPDATE systems SET setup = false WHERE setup AND instance_tenant_id IS NULL;",
    {},
  );
  if (out !== "UPDATE 1") {
    throw new Error(`expected to reopen the instance's setup, got "${out}"`);
  }
  evictCachedSystem();
}

export function closeSetup() {
  const out = sql(
    "UPDATE systems SET setup = true, instance_tenant_id = NULL;",
    {},
  );
  if (out !== "UPDATE 1") {
    throw new Error(`expected to close the instance's setup, got "${out}"`);
  }
  evictCachedSystem();
}
