import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import CreateInstallKeyDrawer from "../CreateInstallKeyDrawer";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockTags } from "@/tests/mockTags";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    installKeyCreate: vi.fn(),
    getTags: vi.fn(),
  }),
);

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

const GENERATED_KEY = "da5132eb-3422-4aca-bb7f-07988e04a636";

async function createKey() {
  const user = userEvent.setup();
  render(
    <ClipboardProvider>
      <CreateInstallKeyDrawer open onClose={vi.fn()} />
    </ClipboardProvider>,
    { wrapper: createTestWrapper() },
  );

  await user.type(screen.getByLabelText("Name"), "fleet-provisioning");
  await user.click(screen.getByRole("button", { name: /create/i }));

  return screen.findByText(/curl -sSf/);
}

describe("CreateInstallKeyDrawer", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedAuthStore({ tenant: "00000000-0000-4000-0000-000000000000" });
    mockTags([]);
    sdk.installKeyCreate.mockResolvedValue(
      mockSdkResponse({ key: GENERATED_KEY }),
    );
  });

  it("installs with the key alone, since the key names its own namespace", async () => {
    const command = await createKey();

    expect(command).toHaveTextContent(`INSTALL_KEY=${GENERATED_KEY}`);
    expect(command).not.toHaveTextContent("TENANT_ID");
  });
});
