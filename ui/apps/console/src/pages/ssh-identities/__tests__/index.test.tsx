import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent, { type UserEvent } from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import SSHIdentities from "../index";
import type { SshIdentity } from "@/client";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockIdentities =
  vi.fn<() => { identities: SshIdentity[]; isLoading: boolean }>();

vi.mock("@/hooks/useSSHIdentities", () => ({
  useSSHIdentities: () => mockIdentities(),
}));

vi.mock("@/hooks/useSSHIdentityMutations", () => ({
  useDeleteSSHIdentity: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (selector: (s: { userId: string }) => unknown) =>
    selector({ userId: "user1" }),
}));

vi.mock("../IdentityDrawer", () => ({ default: () => null }));

const mockBrowserKeyFingerprint = vi.fn<() => string | null>();

vi.mock("@/hooks/useBrowserKey", () => ({
  useBrowserKeyFingerprint: () => mockBrowserKeyFingerprint(),
}));

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

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
  mockIdentities.mockReturnValue({ identities, isLoading: false });

  return render(
    <MemoryRouter>
      <ClipboardProvider>
        <SSHIdentities />
      </ClipboardProvider>
    </MemoryRouter>,
  );
}

function rowFor(name: string) {
  return screen.getByText(name).closest("tr") as HTMLElement;
}

/** Open a row's overflow menu and choose Revoke, the way a person would. */
async function openRevoke(user: UserEvent, name: string) {
  await user.click(
    within(rowFor(name)).getByRole("button", { name: /actions for/i }),
  );
  await user.click(screen.getByRole("menuitem", { name: /revoke/i }));
}

beforeEach(() => {
  vi.clearAllMocks();
  mockBrowserKeyFingerprint.mockReturnValue(null);
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("SSHIdentities", () => {
  // The origin is drawn as a glyph, so the words only survive as the icon's
  // accessible name. That is exactly what a screen reader reads out, which
  // makes it the right thing to assert on.
  it("says where each key came from", () => {
    renderList([
      identity({ id: "a", name: "chrome", source: "browser" }),
      identity({ id: "b", name: "laptop", source: "manual" }),
      identity({ id: "c", name: "workstation", source: "approval" }),
    ]);

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

  // Among several browser keys, the only one a person can act on without going
  // to another machine is the one this browser holds.
  it("picks out the key held by the browser in use", () => {
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

    expect(
      within(rowFor("here")).getByRole("img", { name: "This browser" }),
    ).toBeInTheDocument();
    expect(
      within(rowFor("elsewhere")).getByRole("img", { name: "Browser" }),
    ).toBeInTheDocument();
  });

  // The cell shows a shortened fingerprint, so the full one has to stay
  // reachable — a truncated fingerprint pasted somewhere matches nothing.
  it("shows a shortened fingerprint but keeps the full one available", () => {
    renderList([identity({ name: "chrome" })]);

    const row = rowFor("chrome");
    expect(within(row).queryByText(FINGERPRINT)).not.toBeInTheDocument();
    expect(within(row).getByText(/^SHA256:hHmU2OTPQjhA/)).toBeInTheDocument();
    expect(within(row).getByTitle(FINGERPRINT)).toBeInTheDocument();
  });

  it("tells the truth about what revoking a browser key does", async () => {
    const user = userEvent.setup();
    renderList([identity({ name: "chrome", source: "browser" })]);

    await openRevoke(user, "chrome");

    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveTextContent(/that browser registers a new key/i);
    expect(dialog).not.toHaveTextContent(/needs approval again/i);
  });

  // Revoking the key of the browser you are looking at logs that terminal out,
  // so the warning has to be about you, not about some other browser.
  it("warns in the first person when revoking this browser's own key", async () => {
    const user = userEvent.setup();
    mockBrowserKeyFingerprint.mockReturnValue(FINGERPRINT);
    renderList([identity({ name: "here", source: "browser" })]);

    await openRevoke(user, "here");

    expect(screen.getByRole("dialog")).toHaveTextContent(
      /this browser registers a new key/i,
    );
  });

  it("tells the truth about what revoking a manually added key does", async () => {
    const user = userEvent.setup();
    renderList([identity({ name: "laptop", source: "manual" })]);

    await openRevoke(user, "laptop");

    expect(screen.getByRole("dialog")).toHaveTextContent(
      /needs approval again/i,
    );
  });
});
