import { vi } from "vitest";
import type { Namespace } from "@/client";
import { getNamespaces, getNamespace } from "@/client/sdk.gen";
import { paginatedResponse, mockSdkResponse } from "./sdk";
import { mockNamespace } from "./factories";

export function mockGetNamespaces(
  namespaces: Array<Partial<Namespace>> = [{}],
) {
  vi.mocked(getNamespaces).mockResolvedValue(
    paginatedResponse(namespaces.map((ns) => mockNamespace(ns))),
  );
}

export function mockGetNamespace(overrides: Partial<Namespace> = {}) {
  vi.mocked(getNamespace).mockResolvedValue(
    mockSdkResponse(mockNamespace(overrides)),
  );
}
