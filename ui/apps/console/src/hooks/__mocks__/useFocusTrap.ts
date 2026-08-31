import { vi } from "vitest";

/**
 * Stands in for the focus trap in tests. The real one moves focus into a dialog, which fights
 * Testing Library's own focus handling and makes assertions depend on render order.
 */
export const useFocusTrap = vi.fn();
