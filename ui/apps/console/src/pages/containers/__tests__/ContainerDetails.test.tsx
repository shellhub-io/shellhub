import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import type { Device } from "@/client/model";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import {
  mockContainer as mockContainerFactory,
  mockNamespace,
} from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

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

vi.mock("@/components/common/ActionDialog", () => ({
  default: () => null,
}));

const mockRequestAction = vi.fn();
let capturedOnSuccess: ((action: string) => void) | undefined;

vi.mock("@/hooks/useActionDialog", () => ({
  useActionDialog: (opts?: { onSuccess?: (action: string) => void }) => {
    capturedOnSuccess = opts?.onSuccess;
    return {
      action: undefined,
      actionKey: "closed",
      requestAction: mockRequestAction,
      close: vi.fn(),
      handleSuccess: vi.fn(),
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

function setContainer(overrides: Partial<Device> = {}) {
  server.use(
    http.get("*/api/containers/:uid", () =>
      HttpResponse.json(makeContainer(overrides)),
    ),
  );
}

function renderPage() {
  return render(<ContainerDetails />, {
    wrapper: createTestWrapper({ initialEntries: ["/containers/test-uid"] }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/containers/:uid", () => HttpResponse.json(null)),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.get("*/api/tags", () => HttpResponse.json([])),
    http.put(
      "*/api/containers/:uid",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.post("*/api/tags", () => new HttpResponse(null, { status: 204 })),
    http.post(
      "*/api/containers/:uid/tags/:name",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.delete(
      "*/api/containers/:uid/tags/:name",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
  mockRequestAction.mockReset();
  mockNavigate.mockReset();
  capturedOnSuccess = undefined;
});

describe("ContainerDetails", () => {
  describe("loading and missing states", () => {
    it("renders a spinner while loading", () => {
      server.use(
        http.get("*/api/containers/:uid", () => new Promise(() => {})),
      );
      renderPage();
      expect(
        screen.getByLabelText("Loading container details"),
      ).toBeInTheDocument();
    });

    it("tells the user the container is missing when the query fails", async () => {
      server.use(
        http.get("*/api/containers/:uid", () =>
          HttpResponse.json({}, { status: 404 }),
        ),
      );
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
      setContainer();
    });

    it("renders the container's fields, including the SSHID and online status", async () => {
      renderPage();
      expect(
        await screen.findByRole("heading", { name: "my-container" }),
      ).toBeInTheDocument();
      expect(screen.getByText("aa:bb:cc:dd:ee:ff")).toBeInTheDocument();
      expect(screen.getByText("Alpine Linux 3.19")).toBeInTheDocument();
      expect(
        screen.getByText("my-namespace.my-container@localhost"),
      ).toBeInTheDocument();
      expect(screen.getByText("Online")).toBeInTheDocument();
    });
  });

  it("renders tag names flattened out of the generated tag objects", async () => {
    setContainer({
      tags: [{ name: "production" }, { name: "edge" }],
    } as unknown as Partial<Device>);
    renderPage();
    expect(await screen.findByText("production")).toBeInTheDocument();
    expect(screen.getByText("edge")).toBeInTheDocument();
  });

  it("hides the SSHID banner for a container that is not accepted", async () => {
    setContainer({ status: "pending" });
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
        setContainer({
          status: status as Device["status"],
          online: false,
        });
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
      setContainer({ status: "accepted" });
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
      setContainer();
      renderPage();
      await screen.findByRole("heading", { name: "my-container" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("remove");

      expect(mockNavigate).toHaveBeenCalledWith("/containers");
    });

    it("stays on the page after any other action", async () => {
      setContainer({ status: "pending" });
      renderPage();
      await screen.findByRole("heading", { name: "my-container" });

      expect(capturedOnSuccess).toBeDefined();
      capturedOnSuccess!("accept");

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });
});
