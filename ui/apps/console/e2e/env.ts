export function requireEnv(
  name: string,
  hint = "run 'stack up' first",
): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is not set, ${hint}`);
  }

  return value;
}

export const adminUser = {
  username: requireEnv("E2E_ADMIN_USER"),
  password: requireEnv("E2E_ADMIN_PASSWORD"),
  namespace: requireEnv("E2E_ADMIN_NAMESPACE"),
};

export const edition = (process.env.E2E_EDITION ?? "community") as
  "community" | "enterprise" | "cloud";

export const isCommunity = edition === "community";
export const isEnterprise = edition === "enterprise";
export const isCloud = edition === "cloud";
export const isEnterpriseOrCloud = edition !== "community";
