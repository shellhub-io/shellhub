import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import CreateProvisioningKeyModal from "../CreateProvisioningKeyModal";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";

vi.mock("@/components/common/Modal", async () => ({
  default: (await import("@/tests/mocks")).MockModal,
}));

const GENERATED_KEY = "da5132eb-3422-4aca-bb7f-07988e04a636";

async function createKey(
  fill?: (user: ReturnType<typeof userEvent.setup>) => Promise<void>,
) {
  const user = userEvent.setup();
  render(<CreateProvisioningKeyModal open onClose={vi.fn()} />, {
    wrapper: createTestWrapper(),
  });

  await user.type(screen.getByLabelText("Name"), "fleet-provisioning");
  await fill?.(user);
  await user.click(screen.getByRole("button", { name: /create/i }));

  return screen.findByText(/curl -sSf/);
}

describe("CreateProvisioningKeyModal", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedAuthStore({ tenant: "00000000-0000-4000-0000-000000000000" });
    server.use(
      http.get("*/api/tags", () => jsonWithTotal([])),
      http.post("*/api/namespaces/provisioning-key", () =>
        HttpResponse.json({ key: GENERATED_KEY }),
      ),
    );
  });

  it("installs with the key alone, since the key names its own namespace", async () => {
    const command = await createKey();

    expect(command).toHaveTextContent(`PROVISIONING_KEY=${GENERATED_KEY}`);
    expect(command).not.toHaveTextContent("TENANT_ID");
  });

  function captureCreateBody() {
    const captured: { body?: Record<string, unknown> } = {};
    server.use(
      http.post("*/api/namespaces/provisioning-key", async ({ request }) => {
        captured.body = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ key: GENERATED_KEY });
      }),
    );
    return captured;
  }

  it("sends replace_offline_devices true by default", async () => {
    const captured = captureCreateBody();

    await createKey();

    expect(captured.body?.replace_offline_devices).toBe(true);
  });

  it("sends replace_offline_devices false when the box is cleared", async () => {
    const captured = captureCreateBody();

    await createKey((user) =>
      user.click(
        screen.getByRole("checkbox", { name: /replace offline devices/i }),
      ),
    );

    expect(captured.body?.replace_offline_devices).toBe(false);
  });
});
