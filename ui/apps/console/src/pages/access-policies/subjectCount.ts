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
