import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import IdentityDrawer from "../IdentityDrawer";
import { useHasPermission } from "@/hooks/useHasPermission";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockCreateIdentity = vi.fn();
const mockCreateServiceAccount = vi.fn();

vi.mock("@/hooks/useSSHIdentityMutations", () => ({
  useCreateSSHIdentity: () => ({ mutateAsync: mockCreateIdentity }),
  useRenameSSHIdentity: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/hooks/useServiceAccountMutations", () => ({
  useCreateServiceAccount: () => ({ mutateAsync: mockCreateServiceAccount }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: vi.fn(),
}));

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

const mockUseHasPermission = vi.mocked(useHasPermission);

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function renderDrawer() {
  return render(<IdentityDrawer open editIdentity={null} onClose={vi.fn()} />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

const KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILqk test@host";

beforeEach(() => {
  vi.clearAllMocks();
  mockCreateIdentity.mockResolvedValue({});
  mockCreateServiceAccount.mockResolvedValue({});
  mockUseHasPermission.mockReturnValue(true);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("IdentityDrawer", () => {
  it("enrolls the pasted key for the caller by default", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.type(screen.getByLabelText(/name/i), "laptop");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() =>
      expect(mockCreateIdentity).toHaveBeenCalledWith({
        body: { name: "laptop", data: KEY },
      }),
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
      expect(mockCreateServiceAccount).toHaveBeenCalledWith({
        // A durable key by default: single_use off, no TTL sent.
        body: { name: "ci-bot", data: KEY, single_use: false },
      }),
    );
    expect(mockCreateIdentity).not.toHaveBeenCalled();
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
      expect(mockCreateServiceAccount).toHaveBeenCalledWith({
        body: { name: "ci-bot", data: KEY, single_use: true },
      }),
    );
  });

  it("hides the service-account option without permission", () => {
    mockUseHasPermission.mockReturnValue(false);
    renderDrawer();

    expect(screen.queryByText("A new service account")).not.toBeInTheDocument();
  });
});
