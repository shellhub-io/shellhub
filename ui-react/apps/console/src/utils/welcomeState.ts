const storageKey = (tenantId: string) => `shellhub:welcomed:${tenantId}`;

/** Returns true if the welcome wizard has already been shown for this tenant. */
export function hasSeenWelcome(tenantId: string): boolean {
  try {
    return localStorage.getItem(storageKey(tenantId)) === "true";
  } catch {
    return false;
  }
}

/** Marks the welcome wizard as shown for this tenant. Idempotent. */
export function markWelcomeSeen(tenantId: string): void {
  try {
    localStorage.setItem(storageKey(tenantId), "true");
  } catch {
    // localStorage may be full or unavailable — fail silently
  }
}
