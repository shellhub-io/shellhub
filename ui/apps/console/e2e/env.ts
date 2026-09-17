export function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`${name} is not set, run 'stack up' first`);
  }

  return value;
}

export const adminUser = {
  username: requireEnv("E2E_ADMIN_USER"),
  password: requireEnv("E2E_ADMIN_PASSWORD"),
  namespace: requireEnv("E2E_ADMIN_NAMESPACE"),
};
