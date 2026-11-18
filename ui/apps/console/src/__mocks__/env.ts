import { vi } from "vitest";
import type * as Env from "../env";

const actual = await vi.importActual<typeof Env>("../env");

export type { ClientConfig, Edition } from "../env";

/** The real defaults, so a test baseline cannot drift from production. */
export const defaultConfig = actual.defaultConfig;

/**
 * Stands in for the real loader. Resolves immediately with whatever getConfig returns, so a test
 * never waits on a fetch that will not happen.
 */
export const loadConfig = vi.fn(() => Promise.resolve(getConfig()));

/**
 * The configuration seen by the code under test: the real defaults, community edition with every
 * optional feature off, until a test overrides it.
 */
export const getConfig = vi.fn<() => Env.ClientConfig>(() => defaultConfig);
