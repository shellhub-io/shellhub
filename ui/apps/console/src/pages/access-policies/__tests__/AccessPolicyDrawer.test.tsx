import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  mockAccessPolicy,
  mockNamespace,
  mockServiceAccount,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import AccessPolicyDrawer from "../AccessPolicyDrawer";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    listServiceAccounts: vi.fn(),
    getTags: vi.fn(),
    createAccessPolicy: vi.fn(),
    updateAccessPolicy: vi.fn(),
  }),
);

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.listServiceAccounts.mockResolvedValue(mockSdkResponse([]));
  sdk.getTags.mockResolvedValue(mockSdkResponse([]));
});

describe("AccessPolicyDrawer", () => {
  it("counts the service accounts when editing a role=service policy", async () => {
    sdk.listServiceAccounts.mockResolvedValue(
      mockSdkResponse([
        mockServiceAccount({ id: "sa-1", name: "ci-bot" }),
        mockServiceAccount({ id: "sa-2", name: "deploy-bot" }),
      ]),
    );

    render(
      <AccessPolicyDrawer
        open
        editPolicy={mockAccessPolicy({
          id: "p1",
          name: "bots",
          subject: { type: "role", value: "service" },
        })}
        onClose={vi.fn()}
      />,
      { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
    );

    await waitFor(() =>
      expect(screen.getByText("service")).toHaveTextContent(/service\D*2/),
    );
  });
});
