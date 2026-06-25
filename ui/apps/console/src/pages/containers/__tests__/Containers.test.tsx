import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedContainer } from "@/hooks/useContainers";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useContainers", () => ({
  useContainers: vi.fn(),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "my-ns" } }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-1" }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: [] }) => unknown) =>
    sel({ sessions: [] }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

vi.mock("@/components/common/PageHeader", () => ({
  default: ({
    title,
    children,
  }: {
    title: string;
    children?: React.ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

vi.mock("@/components/common/TagFilterDropdown", () => ({
  default: () => <div />,
}));

vi.mock("@/components/ManageTagsDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("../ContainerTagsPopover", () => ({
  default: ({ container }: { container: NormalizedContainer }) => (
    <span>
      {container.tags.length > 0 ? container.tags.join(", ") : "No tags"}
    </span>
  ),
}));

vi.mock("../ContainerActionDialog", () => ({
  default: () => <div />,
}));

vi.mock("../AddDockerConnectorDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/billing/BillingWarning", () => ({
  default: () => <div />,
}));

vi.mock("@/env", () => ({
  getConfig: () => ({ cloud: false }),
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import React from "react";
import { useContainers } from "@/hooks/useContainers";
import Containers from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  containers: [] as NormalizedContainer[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeContainer(
  overrides: Partial<NormalizedContainer> = {},
): NormalizedContainer {
  return {
    uid: "container-uid-1",
    name: "my-container",
    status: "accepted",
    online: true,
    tags: [],
    last_seen: new Date().toISOString(),
    created_at: new Date().toISOString(),
    info: {
      id: "ubuntu",
      pretty_name: "Ubuntu 22.04 LTS",
      arch: "x86_64",
      platform: "linux",
      version: "0.14.0",
    },
    ...overrides,
  } as NormalizedContainer;
}

function renderPage() {
  return render(
    <MemoryRouter>
      <Containers />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("Containers list", () => {
  beforeEach(() => {
    vi.mocked(useContainers).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Containers" }),
      ).toBeInTheDocument();
    });
  });

  describe("sorting", () => {
    it("requests last_seen/desc sort by default", () => {
      renderPage();
      expect(useContainers).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "last_seen", orderBy: "desc" }),
      );
    });

    it("toggles sort when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [makeContainer({ uid: "uid-1", name: "alpha" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      let calls = vi.mocked(useContainers).mock.calls;
      let last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "asc" });

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      calls = vi.mocked(useContainers).mock.calls;
      last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "desc" });
    });
  });
});
