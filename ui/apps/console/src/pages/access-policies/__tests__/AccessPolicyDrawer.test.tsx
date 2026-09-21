import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import AccessPolicyDrawer from "../AccessPolicyDrawer";

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

let createBody: unknown = null;

beforeEach(() => {
  createBody = null;
  server.use(
    ...defaultHandlers,
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.get("*/api/namespaces/api-key", () => HttpResponse.json([])),
    http.get("*/api/tags", () => HttpResponse.json([])),
    http.post("*/api/access-policies", async ({ request }) => {
      createBody = await request.json();
      return HttpResponse.json({ id: "p1" });
    }),
  );
  seedAuthStore();
});

describe("AccessPolicyDrawer", () => {
  it("saves an API key as the subject, by id", async () => {
    const user = userEvent.setup();

    server.use(
      http.get("*/api/namespaces/api-key", () =>
        HttpResponse.json([
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
      ),
    );

    render(<AccessPolicyDrawer open editPolicy={null} onClose={vi.fn()} />, {
      wrapper: createTestWrapper({ initialEntries: ["/"] }),
    });

    await user.type(screen.getByLabelText(/name/i), "ci reaches prod");
    await user.click(await screen.findByRole("button", { name: /all members/i }));
    await user.click(await screen.findByRole("button", { name: /api keys/i }));
    await user.click(await screen.findByText("ci-deploy"));
    await user.click(screen.getByRole("button", { name: /create|save/i }));

    await waitFor(() => expect(createBody).not.toBeNull());

    expect((createBody as Record<string, unknown>).subject).toEqual({
      type: "api-key",
      value: "c629572a-b643-4301-90fe-4572b00d007e",
    });
  });
});
