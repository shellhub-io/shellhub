import { execFileSync } from "node:child_process";
import { closeSync, existsSync, mkdirSync, openSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const e2eDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = join(e2eDir, "..", "..", "..", "..");
const tmpDir = join(e2eDir, "tmp");
const envFile = join(tmpDir, "stack.env");
const logFile = join(tmpDir, "containers.log");
const composeBin = join(repoRoot, "bin", "docker-compose");
const cloudDir = join(repoRoot, "..", "cloud");

const composeTimeout = 30 * 60 * 1000;
const logsTimeout = 60_000;

const editions = ["community", "enterprise", "cloud"] as const;
type Edition = (typeof editions)[number];

function resolveEdition(): Edition {
  const requested = (process.env.E2E_EDITION ?? "community").trim().toLowerCase();

  if (!editions.includes(requested as Edition)) {
    throw new Error(
      `E2E_EDITION '${requested}' is not one of ${editions.join(", ")}`,
    );
  }

  return requested as Edition;
}

export const edition = resolveEdition();
const buildEdition = edition === "community" ? "community" : "enterprise";
const httpPort = "8090";

const composeFiles = [
  "docker-compose.test.yml",
  "docker-compose.postgres.test.yml",
  "docker-compose.build.test.yml",
  ...(edition === "community" ? [] : ["docker-compose.enterprise.test.yml"]),
];

export const baseURL = `http://localhost:${httpPort}`;

const stackEnv = {
  SHELLHUB_EDITION: edition,
  SHELLHUB_ENV: "production",
  SHELLHUB_BIND_ADDRESS: "127.0.0.1",
  SHELLHUB_DOMAIN: "localhost",
  SHELLHUB_HTTP_PORT: httpPort,
  SHELLHUB_SSH_PORT: "2223",
  SHELLHUB_NETWORK: "shellhub_network_e2e",
  SHELLHUB_MAXMIND_MIRROR: "",
  SHELLHUB_BUILD_EDITION: buildEdition,
  SHELLHUB_CLOUD_SRC: edition === "community" ? "." : "../cloud",
  SHELLHUB_BILLING: "dummy",
  SHELLHUB_EMAIL_PROVIDER: "dummy",
};

const inheritedEnv = Object.fromEntries(
  Object.entries(process.env).filter(([key]) => !key.startsWith("SHELLHUB_")),
);

const composeEnv = {
  ...inheritedEnv,
  ENV_OVERRIDE: envFile,
  COMPOSE_OVERRIDE: "false",
  EXTRA_COMPOSE_FILE: composeFiles.join(":"),
  COMPOSE_PROJECT_NAME: "shellhub-e2e",
};

export function errorReason(error: unknown) {
  if (!(error instanceof Error)) return "unknown failure";

  return error.cause instanceof Error ? error.cause.message : error.message;
}

function writeEnvFile() {
  mkdirSync(tmpDir, { recursive: true });
  writeFileSync(
    envFile,
    Object.entries(stackEnv)
      .map(([key, value]) => `${key}=${value}\n`)
      .join(""),
  );
}

function compose(args: string[]) {
  execFileSync(composeBin, args, {
    cwd: repoRoot,
    stdio: "inherit",
    env: composeEnv,
    timeout: composeTimeout,
  });
}

export function adminCLI(args: string[]) {
  execFileSync(
    composeBin,
    ["exec", "-T", "server", "/server", "admin", ...args],
    { cwd: repoRoot, stdio: ["ignore", "ignore", "inherit"], env: composeEnv },
  );
}

function dumpLogs() {
  let target: number | undefined;

  try {
    mkdirSync(tmpDir, { recursive: true });
    target = openSync(logFile, "w");
    execFileSync(composeBin, ["logs", "--no-color"], {
      cwd: repoRoot,
      stdio: ["ignore", target, target],
      env: composeEnv,
      timeout: logsTimeout,
    });
  } catch (error) {
    process.stderr.write(`container logs unavailable: ${errorReason(error)}\n`);
  } finally {
    if (target !== undefined) closeSync(target);
  }
}

export function up() {
  if (edition !== "community") {
    const required =
      edition === "cloud" ? join(cloudDir, "docker-compose.yml") : cloudDir;

    if (!existsSync(required)) {
      throw new Error(`E2E_EDITION=${edition} requires ${required}`);
    }
  }

  writeEnvFile();
  compose(["down", "-v", "--remove-orphans"]);

  execFileSync("make", ["keygen"], { cwd: repoRoot, stdio: "inherit" });
  compose(["up", "-d", "--build", "--wait", "--wait-timeout", "600"]);
}

export function down() {
  dumpLogs();

  try {
    compose(["down", "-v", "--remove-orphans"]);
  } catch (error) {
    process.stderr.write(`teardown failed: ${errorReason(error)}\n`);
  }
}
