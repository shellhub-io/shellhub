import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  mockNamespace,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import AccessPolicyDrawer from "../AccessPolicyDrawer";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    getTags: vi.fn(),
    apiKeyList: vi.fn(),
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
  sdk.apiKeyList.mockResolvedValue(mockSdkResponse([]));
  sdk.getTags.mockResolvedValue(mockSdkResponse([]));
});

describe("AccessPolicyDrawer", () => {
  it("saves an API key as the subject, by id", async () => {
    const user = userEvent.setup();

    sdk.apiKeyList.mockResolvedValue(
      mockSdkResponse([
        {
          id: "c629572a-b643-4301-90fe-4572b00d007e",
          name: "ci-deploy",
          tenant_id: "00000000-0000-4000-0000-000000000000",
          created_by: "user-1",
          role: "administrator",
          expires_in: -1,
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        },
      ]),
    );
    sdk.createAccessPolicy.mockResolvedValue(mockSdkResponse({ id: "p1" }));

    render(<AccessPolicyDrawer open editPolicy={null} onClose={vi.fn()} />, {
      wrapper: createTestWrapper({ initialEntries: ["/"] }),
    });

    await user.type(screen.getByLabelText(/name/i), "ci reaches prod");
    await user.click(await screen.findByRole("button", { name: /all members/i }));
    await user.click(await screen.findByRole("button", { name: /api keys/i }));
    await user.click(await screen.findByText("ci-deploy"));
    await user.click(screen.getByRole("button", { name: /create|save/i }));

    await waitFor(() => expect(sdk.createAccessPolicy).toHaveBeenCalled());

    const body = sdk.createAccessPolicy.mock.calls[0][0].body;
    expect(body.subject).toEqual({
      type: "api-key",
      value: "c629572a-b643-4301-90fe-4572b00d007e",
    });
  });
});
