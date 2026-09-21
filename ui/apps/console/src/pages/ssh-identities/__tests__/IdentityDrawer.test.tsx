import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import IdentityDrawer from "../IdentityDrawer";
import { useAuthStore } from "@/stores/authStore";

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

let identityCalled: boolean;

beforeEach(() => {
  vi.clearAllMocks();
  identityCalled = false;
  useAuthStore.setState({ role: "owner" });
  server.use(
    http.post("*/api/ssh-identities", () => {
      identityCalled = true;
      return HttpResponse.json({});
    }),
    http.patch(
      "*/api/ssh-identities/:id",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("IdentityDrawer", () => {
  it("enrolls the pasted key for the caller by default", async () => {
    const user = userEvent.setup();
    renderDrawer();

    await user.type(screen.getByLabelText(/name/i), "laptop");
    await user.type(screen.getByLabelText(/public key data/i), KEY);
    await user.click(screen.getByRole("button", { name: /add key/i }));

    await waitFor(() => expect(identityCalled).toBe(true));
  });
});
