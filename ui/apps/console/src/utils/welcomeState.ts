const storageKey = (tenantId: string) => `shellhub:welcomed:${tenantId}`;

/** Returns true if the welcome wizard has already been shown for this tenant. */
export function hasSeenWelcome(tenantId: string): boolean {
  try {
    return localStorage.getItem(storageKey(tenantId)) === "true";
  } catch {
    return false;
  }
}

/**
 * Marks the welcome wizard as shown for this tenant, reporting whether it stuck. Idempotent.
 * Storage can be full or blocked, and the flag is not worth failing over — a caller that does
 * not check simply risks showing the wizard again.
 */
export function markWelcomeSeen(tenantId: string): boolean {
  try {
    localStorage.setItem(storageKey(tenantId), "true");
    return true;
  } catch {
    return false;
  }
}
