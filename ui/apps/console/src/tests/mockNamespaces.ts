import { vi } from "vitest";
import type { Namespace } from "@/client";
import { getNamespaces, getNamespace } from "@/client/sdk.gen";
import { paginatedResponse, mockSdkResponse } from "./sdk";
import { mockNamespace } from "./factories";

/**
 * Makes the namespaces list resolve to the given namespaces, filled out from the factory.
 */
export function mockGetNamespaces(
  namespaces: Array<Partial<Namespace>> = [{}],
) {
  vi.mocked(getNamespaces).mockResolvedValue(
    paginatedResponse(namespaces.map((ns) => mockNamespace(ns))),
  );
}

/**
 * Makes the single-namespace call resolve to one built from the factory.
 */
export function mockGetNamespace(overrides: Partial<Namespace> = {}) {
  vi.mocked(getNamespace).mockResolvedValue(
    mockSdkResponse(mockNamespace(overrides)),
  );
}
