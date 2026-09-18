import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import IdentityDrawer from "../IdentityDrawer";
import { useAuthStore } from "@/stores/authStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    createSshIdentity: vi.fn(),
    renameSshIdentity: vi.fn(),
    createApiKeySshIdentity: vi.fn(),
    apiKeyList: vi.fn(),
  }),
);

vi.mock("@/utils/sshKeys", () => ({
  isPublicKeyValid: () => true,
}));

vi.mock("@/components/common/Drawer", async () => ({
  default: (await import("@/tests/mocks")).MockDrawer,
}));

vi.mock("@/components/common/fields/KeyFileInput", () => ({
  default: ({
    label,
    value,
    onChange,
  }: {
    label: string;
    value: string;
    onChange: (v: string) => void;
  }) => (
    <textarea
      aria-label={label}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

function renderDrawer() {
  return render(<IdentityDrawer open editIdentity={null} onClose={vi.fn()} />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

const KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILqk test@host";

beforeEach(() => {
  vi.clearAllMocks();
  sdk.createSshIdentity.mockResolvedValue(mockSdkResponse({}));
  sdk.createApiKeySshIdentity.mockResolvedValue(mockSdkResponse({}));
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
  useAuthStore.setState({ role: "owner" });
});

describe("IdentityDrawer", () => {
  it("enrolls the pasted key for the caller by default", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.type(screen.getByLabelText("Name"), "laptop");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() =>
      expect(sdk.createSshIdentity).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { name: "laptop", data: KEY },
        }),
      ),
    );
    expect(sdk.createApiKeySshIdentity).not.toHaveBeenCalled();
  });

  it("enrols the key for the chosen API key, by name", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.click(screen.getByText("An API key"));
    await user.selectOptions(screen.getByRole("combobox"), "ci-deploy");
    await user.type(screen.getByLabelText("Name"), "deploy");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() =>
      expect(sdk.createApiKeySshIdentity).toHaveBeenCalledWith(
        expect.objectContaining({
          path: { name: "ci-deploy" },
          body: { name: "deploy", data: KEY, single_use: false },
        }),
      ),
    );
    expect(sdk.createSshIdentity).not.toHaveBeenCalled();
  });

  it("burns the key after one session when the toggle is on", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.click(screen.getByText("An API key"));
    await user.selectOptions(screen.getByRole("combobox"), "ci-deploy");
    await user.type(screen.getByLabelText("Name"), "deploy");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("switch", { name: /single-use key/i }));
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() =>
      expect(sdk.createApiKeySshIdentity).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ single_use: true }),
        }),
      ),
    );
  });

  it("hides the API key option without permission", () => {
    useAuthStore.setState({ role: "observer" });
    renderDrawer();

    expect(screen.queryByText("An API key")).not.toBeInTheDocument();
  });
});
