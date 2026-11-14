/**
 * Whether a session token can no longer be used: its `exp` claim is in the past, or it is not a
 * JWT whose payload can be read, which the server would refuse anyway.
 */
export function isTokenExpired(token: string): boolean {
  try {
    const payload: unknown = JSON.parse(atob(token.split(".")[1]));
    if (typeof payload === "object" && payload !== null && "exp" in payload) {
      const { exp } = payload;
      return typeof exp === "number" && exp * 1000 < Date.now();
    }
    return false;
  } catch {
    return true;
  }
}

/**
 * Whether the login page was opened to finish or start a sign-in the server sent it, so a stale
 * session in the store must not send the page away: `?token=` from "Login as User", `?sso_code=`
 * from a finished SSO sign-in, or `?sso=start` from a sign-in the IdP started.
 */
export function isSignInLanding(search: string): boolean {
  const params = new URLSearchParams(search);
  return (
    params.has("token") ||
    params.has("sso_code") ||
    params.get("sso") === "start"
  );
}
