import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import React from "react";
import type { Device } from "@/client";

vi.mock("@/hooks/useContainer", () => ({
  useContainer: vi.fn(),
}));

vi.mock("@/hooks/useContainerMutations", () => ({
  useRenameContainer: () => ({ mutateAsync: vi.fn() }),
  useAddContainerTag: () => ({ mutateAsync: vi.fn() }),
  useRemoveContainerTag: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "my-ns" } }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-1" }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (
    sel: (s: { sessions: []; restore: () => void }) => unknown,
  ) => sel({ sessions: [], restore: vi.fn() }),
}));

// TagsSection, the one child that fetches, reaches for both of these.
vi.mock("@/hooks/useTags", () => ({
  useTags: () => ({ tags: [], totalCount: 0, isLoading: false, error: null }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock("@/pages/containers/ContainerActionsPortal", () => ({
  default: () => null,
}));

const mockRequestAction = vi.fn();
let capturedOnSuccess: ((action: string) => void) | undefined;

vi.mock("@/hooks/useContainerActions", () => ({
  useContainerActions: (opts?: { onSuccess?: (action: string) => void }) => {
    capturedOnSuccess = opts?.onSuccess;
    return {
      operation: undefined,
      requestAction: mockRequestAction,
      close: vi.fn(),
      billingWarningOpen: false,
      closeBillingWarning: vi.fn(),
      onBillingWarning: undefined,
      runSuccess: vi.fn(),
      billingEnabled: false,
    };
  },
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

const mockNavigate = vi.fn();

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return {
    ...actual,
    useParams: () => ({ uid: "test-uid" }),
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  };
});

import { useContainer } from "@/hooks/useContainer";
import ContainerDetails from "@/pages/ContainerDetails";

function makeContainer(overrides: Partial<Device> = {}): Device {
  return {
    uid: "test-uid",
    name: "my-container",
    status: "accepted",
    online: true,
    tags: [],
    last_seen: "2024-01-15T10:00:00.000Z",
    created_at: "2023-06-01T08:00:00.000Z",
    status_updated_at: "2023-06-02T08:00:00.000Z",
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "alpine",
      pretty_name: "Alpine Linux 3.19",
      arch: "x86_64",
      platform: "docker",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    ...overrides,
  } as Device;
}

function mockContainer(
  state: Partial<{
    container: Device | null;
    isLoading: boolean;
    error: unknown;
  }> = {},
) {
  vi.mocked(useContainer).mockReturnValue({
    container: null,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
    ...state,
  } as unknown as ReturnType<typeof useContainer>);
}

function renderPage() {
  return render(
    <MemoryRouter>
      <ContainerDetails />
    </MemoryRouter>,
  );
}

describe("ContainerDetails", () => {
  beforeEach(() => {
    mockRequestAction.mockReset();
    mockNavigate.mockReset();
    capturedOnSuccess = undefined;
    mockContainer();
  });

  describe("loading and missing states", () => {
    it("renders a spinner while loading", () => {
      mockContainer({ isLoading: true });
      renderPage();
      expect(screen.getByLabelText("Loading container details")).toBeInTheDocument();
    });

    it("tells the user the container is missing when the query fails", () => {
      mockContainer({ error: new Error("boom") });
      renderPage();
      expect(screen.getByText("Container not found")).toBeInTheDocument();
    });

    it("tells the user the container is missing when the query returns none", () => {
      mockContainer({ container: null });
      renderPage();
      expect(screen.getByText("Container not found")).toBeInTheDocument();
    });
  });

  describe("container data", () => {
    beforeEach(() => {
      mockContainer({ container: makeContainer() });
    });

    it("renders the container name as a heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "my-container" }),
      ).toBeInTheDocument();
    });

    it("renders the MAC address", () => {
      renderPage();
      expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
    });

    it("renders the container image", () => {
      renderPage();
      expect(screen.getByText("Alpine Linux 3.19")).toBeInTheDocument();
    });

    it("renders the SSHID built from the namespace and container name", () => {
      renderPage();
      expect(
        screen.getByText("my-ns.my-container@localhost"),
      ).toBeInTheDocument();
    });

    it("marks an online container as Online", () => {
      renderPage();
      expect(screen.getByText("Online")).toBeInTheDocument();
    });
  });

  it("renders tag names flattened out of the generated tag objects", () => {
    mockContainer({
      container: makeContainer({
        tags: [{ name: "production" }, { name: "edge" }],
      } as unknown as Partial<Device>),
    });
    renderPage();

    expect(screen.getByText("production")).toBeInTheDocument();
    expect(screen.getByText("edge")).toBeInTheDocument();
  });

  it("hides the SSHID banner for a container that is not accepted", () => {
    mockContainer({ container: makeContainer({ status: "pending" }) });
    renderPage();

    expect(
      screen.queryByText("my-ns.my-container@localhost"),
    ).not.toBeInTheDocument();
  });

  describe("action buttons delegate to useContainerActions", () => {
    it.each([
      ["pending", /Accept/i, "accept"],
      ["pending", /Reject/i, "reject"],
      ["rejected", /Accept/i, "accept"],
      ["rejected", /Remove/i, "remove"],
    ])(
      "calls requestAction('%s' → %s)",
      async (status, buttonName, expectedAction) => {
        const user = userEvent.setup();
        mockContainer({
          container: makeContainer({
            status: status as Device["status"],
            online: false,
          }),
        });
        renderPage();

        await user.click(screen.getByRole("button", { name: buttonName }));

        expect(mockRequestAction).toHaveBeenCalledWith(
          expect.objectContaining({ uid: "test-uid" }),
          expectedAction,
        );
      },
    );

    it("calls requestAction('remove') from the trash button on an accepted container", async () => {
      const user = userEvent.setup();
      mockContainer({ container: makeContainer({ status: "accepted" }) });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Remove container" }),
      );

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "remove",
      );
    });
  });

  describe("onSuccess callback wiring", () => {
    it("navigates to /containers after a container is removed", () => {
      mockContainer({ container: makeContainer() });
      renderPage();

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("remove");

      expect(mockNavigate).toHaveBeenCalledWith("/containers");
    });

    it("stays on the page after any other action", () => {
      mockContainer({ container: makeContainer({ status: "pending" }) });
      renderPage();

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("accept");

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });
});
