import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import type { Device } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import {
  mockContainer as mockContainerFactory,
  mockNamespace,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getContainer: vi.fn(),
    updateContainer: vi.fn(),
    createTag: vi.fn(),
    pushTagToContainer: vi.fn(),
    pullTagFromContainer: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    getTags: vi.fn(),
  }),
);

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (
    sel: (s: { sessions: []; restore: () => void }) => unknown,
  ) => sel({ sessions: [], restore: vi.fn() }),
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

import ContainerDetails from "@/pages/ContainerDetails";

function makeContainer(overrides: Partial<Device> = {}): Device {
  return mockContainerFactory({
    uid: "test-uid",
    name: "my-container",
    status: "accepted",
    online: true,
    last_seen: "2024-01-15T10:00:00.000Z",
    created_at: "2023-06-01T08:00:00.000Z",
    info: {
      id: "alpine",
      pretty_name: "Alpine Linux 3.19",
      arch: "x86_64",
      platform: "docker",
      version: "0.14.0",
    },
    ...overrides,
  });
}

function renderPage() {
  return render(<ContainerDetails />, {
    wrapper: createTestWrapper({ initialEntries: ["/containers/test-uid"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getContainer.mockResolvedValue(mockSdkResponse(null));
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.getTags.mockResolvedValue(mockSdkResponse([]));
  sdk.updateContainer.mockResolvedValue(mockSdkResponse(undefined));
  sdk.createTag.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pushTagToContainer.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pullTagFromContainer.mockResolvedValue(mockSdkResponse(undefined));
  mockRequestAction.mockReset();
  mockNavigate.mockReset();
  capturedOnSuccess = undefined;
});

describe("ContainerDetails", () => {
  describe("loading and missing states", () => {
    it("renders a spinner while loading", () => {
      sdk.getContainer.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(
        screen.getByLabelText("Loading container details"),
      ).toBeInTheDocument();
    });

    it("tells the user the container is missing when the query fails", async () => {
      sdk.getContainer.mockRejectedValue({ status: 404 });
      renderPage();
      expect(
        await screen.findByText("Container not found"),
      ).toBeInTheDocument();
    });

    it("tells the user the container is missing when the query returns none", async () => {
      renderPage();
      expect(
        await screen.findByText("Container not found"),
      ).toBeInTheDocument();
    });
  });

  describe("container data", () => {
    beforeEach(() => {
      sdk.getContainer.mockResolvedValue(mockSdkResponse(makeContainer()));
    });

    it("renders the container name as a heading", async () => {
      renderPage();
      expect(
        await screen.findByRole("heading", { name: "my-container" }),
      ).toBeInTheDocument();
    });

    it("renders the MAC address", async () => {
      renderPage();
      expect(await screen.findByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
    });

    it("renders the container image", async () => {
      renderPage();
      expect(await screen.findByText("Alpine Linux 3.19")).toBeInTheDocument();
    });

    it("renders the SSHID built from the namespace and container name", async () => {
      renderPage();
      expect(
        await screen.findByText("my-namespace.my-container@localhost"),
      ).toBeInTheDocument();
    });

    it("marks an online container as Online", async () => {
      renderPage();
      expect(await screen.findByText("Online")).toBeInTheDocument();
    });
  });

  it("renders tag names flattened out of the generated tag objects", async () => {
    sdk.getContainer.mockResolvedValue(
      mockSdkResponse(
        makeContainer({
          tags: [{ name: "production" }, { name: "edge" }],
        } as unknown as Partial<Device>),
      ),
    );
    renderPage();
    expect(await screen.findByText("production")).toBeInTheDocument();
    expect(screen.getByText("edge")).toBeInTheDocument();
  });

  it("hides the SSHID banner for a container that is not accepted", async () => {
    sdk.getContainer.mockResolvedValue(
      mockSdkResponse(makeContainer({ status: "pending" })),
    );
    renderPage();
    await screen.findByRole("heading", { name: "my-container" });
    expect(
      screen.queryByText("my-namespace.my-container@localhost"),
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
        sdk.getContainer.mockResolvedValue(
          mockSdkResponse(
            makeContainer({
              status: status as Device["status"],
              online: false,
            }),
          ),
        );
        renderPage();

        await user.click(
          await screen.findByRole("button", { name: buttonName }),
        );

        expect(mockRequestAction).toHaveBeenCalledWith(
          expect.objectContaining({ uid: "test-uid" }),
          expectedAction,
        );
      },
    );

    it("calls requestAction('remove') from the trash button on an accepted container", async () => {
      const user = userEvent.setup();
      sdk.getContainer.mockResolvedValue(
        mockSdkResponse(makeContainer({ status: "accepted" })),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", { name: "Remove container" }),
      );

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "test-uid" }),
        "remove",
      );
    });
  });

  describe("onSuccess callback wiring", () => {
    it("navigates to /containers after a container is removed", async () => {
      sdk.getContainer.mockResolvedValue(mockSdkResponse(makeContainer()));
      renderPage();
      await screen.findByRole("heading", { name: "my-container" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("remove");

      expect(mockNavigate).toHaveBeenCalledWith("/containers");
    });

    it("stays on the page after any other action", async () => {
      sdk.getContainer.mockResolvedValue(
        mockSdkResponse(makeContainer({ status: "pending" })),
      );
      renderPage();
      await screen.findByRole("heading", { name: "my-container" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("accept");

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });
});
