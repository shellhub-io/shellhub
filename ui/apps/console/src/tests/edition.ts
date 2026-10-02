import { vi } from "vitest";
import { getConfig, defaultConfig, type Edition } from "@/env";

/**
 * Runs the code under test as the given edition, every other setting at its test default. The
 * edition sticks until the next call, so a file that switches it resets it in beforeEach.
 */
export function setEdition(edition: Edition) {
  vi.mocked(getConfig).mockReturnValue({ ...defaultConfig, edition });
}
