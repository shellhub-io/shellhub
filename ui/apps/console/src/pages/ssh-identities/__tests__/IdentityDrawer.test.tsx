import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import IdentityDrawer from "../IdentityDrawer";
import { useAuthStore } from "@/stores/authStore";

const mockCreateSshIdentity = vi.hoisted(() => vi.fn());
const mockRenameSshIdentity = vi.hoisted(() => vi.fn());
const mockCreateServiceAccount = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    createSshIdentity: mockCreateSshIdentity,
    renameSshIdentity: mockRenameSshIdentity,
    createServiceAccount: mockCreateServiceAccount,
  };
});

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
  mockCreateSshIdentity.mockResolvedValue(mockSdkResponse({}));
  mockCreateServiceAccount.mockResolvedValue(mockSdkResponse({}));
  useAuthStore.setState({ role: "owner" });
});

describe("IdentityDrawer", () => {
  it("enrolls the pasted key for the caller by default", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.type(screen.getByLabelText(/name/i), "laptop");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() =>
      expect(mockCreateSshIdentity).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { name: "laptop", data: KEY },
        }),
      ),
    );
    expect(mockCreateServiceAccount).not.toHaveBeenCalled();
  });

  it("creates a service account when that target is chosen", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.click(screen.getByText("A new service account"));
    await user.type(screen.getByLabelText(/name/i), "ci-bot");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(
      screen.getByRole("button", { name: /create service account/i }),
    );

    await waitFor(() =>
      expect(mockCreateServiceAccount).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { name: "ci-bot", data: KEY, single_use: false },
        }),
      ),
    );
    expect(mockCreateSshIdentity).not.toHaveBeenCalled();
  });

  it("creates a single-use service account when the toggle is on", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.click(screen.getByText("A new service account"));
    await user.type(screen.getByLabelText(/name/i), "ci-bot");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("switch", { name: /single-use key/i }));
    await user.click(
      screen.getByRole("button", { name: /create service account/i }),
    );

    await waitFor(() =>
      expect(mockCreateServiceAccount).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { name: "ci-bot", data: KEY, single_use: true },
        }),
      ),
    );
  });

  it("hides the service-account option without permission", () => {
    useAuthStore.setState({ role: "observer" });
    renderDrawer();

    expect(screen.queryByText("A new service account")).not.toBeInTheDocument();
  });
});
