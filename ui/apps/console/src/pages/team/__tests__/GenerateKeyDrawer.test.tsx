import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import GenerateKeyDrawer from "../GenerateKeyDrawer";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { createTestWrapper } from "@/tests/wrapper";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

const SURROGATE_ID = "c629572a-b643-4301-90fe-4572b00d007e";
const CREDENTIAL = "2f6a1d8e-5b0c-4e77-9a3f-0c1d2e3f4a5b";

describe("GenerateKeyDrawer", () => {
  beforeEach(() => {
    server.use(
      ...defaultHandlers,
      http.post("*/api/namespaces/api-key", () =>
        HttpResponse.json({
          id: SURROGATE_ID,
          key: CREDENTIAL,
          tenant_id: "00000000-0000-4000-0000-000000000000",
          created_by: "user-123",
          role: "administrator",
          name: "prod-key",
          expires_in: 0,
          created_at: "2024-01-01T00:00:00Z",
          updated_at: "2024-01-01T00:00:00Z",
        }),
      ),
    );
    seedAuthStore();
  });

  it("shows the credential, never the identifier", async () => {
    const user = userEvent.setup();

    render(
      <ClipboardProvider>
        <GenerateKeyDrawer open onClose={vi.fn()} />
      </ClipboardProvider>,
      { wrapper: createTestWrapper() },
    );

    await user.type(screen.getByLabelText("Name"), "prod-key");
    await user.click(screen.getByRole("button", { name: /generate key/i }));

    expect(await screen.findByText(CREDENTIAL)).toBeInTheDocument();
    expect(screen.queryByText(SURROGATE_ID)).not.toBeInTheDocument();
  });
});
