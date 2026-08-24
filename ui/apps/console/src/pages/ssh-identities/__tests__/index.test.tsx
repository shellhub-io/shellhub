import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent, { type UserEvent } from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import SSHIdentities from "../index";
import type { SshIdentity } from "@/client";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { mockSdkResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listSshIdentities: vi.fn(),
  }),
);

vi.mock("../IdentityDrawer", () => ({ default: () => null }));

const mockBrowserKeyFingerprint = vi.fn<() => string | null>();

vi.mock("@/hooks/useBrowserKey", () => ({
  useBrowserKeyFingerprint: () => mockBrowserKeyFingerprint(),
}));

const FINGERPRINT = "SHA256:hHmU2OTPQjhAKm3ecpf4iw3lqWNCWaFbG1kBje0kn0";

function identity(overrides: Partial<SshIdentity> = {}): SshIdentity {
  return {
    id: "id1",
    principal_id: "user1",
    principal_name: "John Doe",
    principal_email: "john@example.com",
    principal_type: "human",
    fingerprint: FINGERPRINT,
    name: "Chrome 149 on Linux",
    source: "browser",
    created_at: "2026-01-01T00:00:00.000Z",
    last_used_at: null,
    single_use: false,
    ...overrides,
  };
}

function renderList(identities: SshIdentity[]) {
  sdk.listSshIdentities.mockResolvedValue(mockSdkResponse(identities));

  return render(
    <MemoryRouter>
      <ClipboardProvider>
        <SSHIdentities />
      </ClipboardProvider>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

function rowFor(name: string) {
  return screen.getByText(name).closest("tr") as HTMLElement;
}

async function openRevoke(user: UserEvent, name: string) {
  await user.click(
    within(rowFor(name)).getByRole("button", { name: /actions for/i }),
  );
  await user.click(screen.getByRole("menuitem", { name: /revoke/i }));
}

beforeEach(() => {
  vi.clearAllMocks();
  mockBrowserKeyFingerprint.mockReturnValue(null);
  useAuthStore.setState({ userId: "user1" });
  sdk.listSshIdentities.mockResolvedValue(mockSdkResponse([]));
});

describe("SSHIdentities", () => {
  it("says where each key came from", async () => {
    renderList([
      identity({ id: "a", name: "chrome", source: "browser" }),
      identity({ id: "b", name: "laptop", source: "manual" }),
      identity({ id: "c", name: "workstation", source: "approval" }),
    ]);

    await screen.findByText("chrome");

    expect(
      within(rowFor("chrome")).getByRole("img", { name: "Browser" }),
    ).toBeInTheDocument();
    expect(
      within(rowFor("laptop")).getByRole("img", { name: "Manual" }),
    ).toBeInTheDocument();
    expect(
      within(rowFor("workstation")).getByRole("img", { name: "At login" }),
    ).toBeInTheDocument();
  });

  it("picks out the key held by the browser in use", async () => {
    mockBrowserKeyFingerprint.mockReturnValue(FINGERPRINT);
    renderList([
      identity({ id: "a", name: "here", source: "browser" }),
      identity({
        id: "b",
        name: "elsewhere",
        source: "browser",
        fingerprint: "SHA256:someotherbrowserkeyvalue",
      }),
    ]);

    await screen.findByText("here");

    expect(
      within(rowFor("here")).getByRole("img", { name: "This browser" }),
    ).toBeInTheDocument();
    expect(
      within(rowFor("elsewhere")).getByRole("img", { name: "Browser" }),
    ).toBeInTheDocument();
  });

  it("shows a shortened fingerprint but keeps the full one available", async () => {
    renderList([identity({ name: "chrome" })]);

    await screen.findByText("chrome");

    const row = rowFor("chrome");
    expect(within(row).queryByText(FINGERPRINT)).not.toBeInTheDocument();
    expect(within(row).getByText(/^SHA256:hHmU2OTPQjhA/)).toBeInTheDocument();
    expect(within(row).getByTitle(FINGERPRINT)).toBeInTheDocument();
  });

  it("tells the truth about what revoking a browser key does", async () => {
    const user = userEvent.setup();
    renderList([identity({ name: "chrome", source: "browser" })]);

    await screen.findByText("chrome");
    await openRevoke(user, "chrome");

    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveTextContent(/that browser registers a new key/i);
    expect(dialog).not.toHaveTextContent(/needs approval again/i);
  });

  it("warns in the first person when revoking this browser's own key", async () => {
    const user = userEvent.setup();
    mockBrowserKeyFingerprint.mockReturnValue(FINGERPRINT);
    renderList([identity({ name: "here", source: "browser" })]);

    await screen.findByText("here");
    await openRevoke(user, "here");

    expect(screen.getByRole("dialog")).toHaveTextContent(
      /this browser registers a new key/i,
    );
  });

  it("tells the truth about what revoking a manually added key does", async () => {
    const user = userEvent.setup();
    renderList([identity({ name: "laptop", source: "manual" })]);

    await screen.findByText("laptop");
    await openRevoke(user, "laptop");

    expect(screen.getByRole("dialog")).toHaveTextContent(
      /needs approval again/i,
    );
  });
});
