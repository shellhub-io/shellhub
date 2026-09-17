/** The membership role a service account carries. It is stamped at creation and never chosen. */
export const SERVICE_ROLE = "service";

/**
 * Counts the principals a role subject matches. The namespace payload holds people only, so a
 * service role is counted from the service accounts instead of from the members.
 */
export function roleSubjectCount({
  role,
  members,
  serviceAccounts,
}: {
  role: string;
  members: { role?: string }[];
  serviceAccounts: unknown[];
}): number {
  return role === SERVICE_ROLE
    ? serviceAccounts.length
    : members.filter((m) => String(m.role) === role).length;
}
