/**
 * Where an account stands. awaiting_approval is not the same as unconfirmed: the person may
 * have verified their address and still be waiting for an admin.
 */
export type UserStatus = "confirmed" | "not-confirmed" | "awaiting_approval";

/**
 * Where an account stands, from the user the admin API returns: awaiting approval overrides the
 * confirmation status beside it.
 */
export function userStatus(user: {
  status: Exclude<UserStatus, "awaiting_approval">;
  awaiting_approval?: boolean;
}): UserStatus {
  return user.awaiting_approval ? "awaiting_approval" : user.status;
}
