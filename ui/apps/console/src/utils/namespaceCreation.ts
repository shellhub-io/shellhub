import { isCommunity } from "@/env";

/**
 * Whether the signed-in account may create a namespace: never on community, where the server has
 * no route for it, and elsewhere only once the account info has loaded with an allowance other
 * than zero. An account still loading counts as not allowed. The allowance is a cap on owned
 * namespaces, and this does not compare it with how many the account already owns: an account at
 * its limit still gets true, and the server then refuses the creation as limit reached.
 */
export function canCreateNamespace(
  user: { max_namespaces: number } | undefined,
): boolean {
  return !isCommunity() && !!user && user.max_namespaces !== 0;
}
