/**
 * The roles a policy may name, which is every role that can connect to a device. Observer
 * cannot, and the server refuses a policy naming it.
 */
export const POLICY_SUBJECT_ROLES = ["owner", "administrator", "operator"] as const;

/**
 * Counts the principals a role subject matches. Only a person holds a role: an automation is an
 * API key, which a policy names directly.
 */
export function roleSubjectCount({
  role,
  members,
}: {
  role: string;
  members: { role?: string }[];
}): number {
  return members.filter((m) => String(m.role) === role).length;
}
