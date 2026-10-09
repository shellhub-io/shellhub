/**
 * What a page hands the sign-in page through the route state: a notice to show above the form,
 * and whether the sign-in page must end the current session first.
 */
export interface LoginState {
  notice?: string;
  signOut: boolean;
}

/**
 * The route state for leaving a page behind ProtectedRoute for the sign-in page, which then
 * ends the session and shows notice.
 */
export function signOutLoginState(notice: string): LoginState {
  return { notice, signOut: true };
}

/**
 * Reads the hand-over from a route's state. A missing or malformed state reads as no notice and
 * no sign-out.
 */
export function readLoginState(state: unknown): LoginState {
  if (typeof state !== "object" || state === null) return { signOut: false };
  const { notice, signOut } = state as Partial<
    Record<keyof LoginState, unknown>
  >;
  return {
    notice: typeof notice === "string" ? notice : undefined,
    signOut: signOut === true,
  };
}
