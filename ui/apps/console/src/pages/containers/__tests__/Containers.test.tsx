import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockContainer, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: [] }) => unknown) =>
    sel({ sessions: [] }),
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
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

vi.mock("@/components/common/TagsPopover", () => ({
  default: ({ tags }: { tags: string[] }) => (
    <span>{tags.length > 0 ? tags.join(", ") : "No tags"}</span>
  ),
}));

vi.mock("@/components/common/ActionDialog", () => ({ default: () => null }));

vi.mock("../AddDockerConnectorDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const mockRequestAction = vi.fn();
vi.mock("@/hooks/useActionDialog", () => ({
  useActionDialog: () => ({
    requestAction: mockRequestAction,
    action: undefined,
    actionKey: "closed",
    close: vi.fn(),
    handleSuccess: vi.fn(),
  }),
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

import Containers from "../index";

function renderPage(initialEntries: string[] = ["/"]) {
  return render(<Containers />, {
    wrapper: createTestWrapper({ initialEntries }),
  });
}

let lastContainersUrl: URL | null;

function setContainers(
  containers: ReturnType<typeof mockContainer>[],
  total?: number,
) {
  server.use(
    http.get("*/api/containers", ({ request }) => {
      lastContainersUrl = new URL(request.url);
      return jsonWithTotal(containers, total ?? containers.length);
    }),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  lastContainersUrl = null;
  seedAuthStore();
  setContainers([]);
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
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
  mockNavigate.mockReset();
  mockRequestAction.mockReset();
});

describe("Containers list", () => {
  it("navigates to container detail on row click", async () => {
    const user = userEvent.setup();
    setContainers([mockContainer({ uid: "uid-abc", name: "clickable" })], 1);
    renderPage();
    await user.click(await screen.findByText("clickable"));
    expect(mockNavigate).toHaveBeenCalledWith("/containers/uid-abc");
  });

  it("honours a non-accepted status from the URL", async () => {
    renderPage(["/?status=pending"]);
    await waitFor(() => {
      expect(lastContainersUrl?.searchParams.get("status")).toBe("pending");
    });
  });

  it("resets page to 1 when a status tab is clicked while on page 2", async () => {
    const user = userEvent.setup();
    renderPage(["/?page=2"]);
    await waitFor(() => {
      expect(lastContainersUrl?.searchParams.get("page")).toBe("2");
    });

    await user.click(screen.getByRole("tab", { name: "Pending" }));

    await waitFor(() => {
      expect(lastContainersUrl!.searchParams.get("page")).toBe("1");
      expect(lastContainersUrl!.searchParams.get("status")).toBe("pending");
    });
  });

  it.each([
    ["pending", "Accept", "accept"],
    ["pending", "Reject", "reject"],
    ["rejected", "Remove", "remove"],
  ] as const)(
    "%s view: clicking %s requests that action for the container",
    async (status, buttonName, expectedAction) => {
      const user = userEvent.setup();
      setContainers(
        [
          mockContainer({
            uid: "uid-1",
            name: "a-container",
            status,
            online: false,
          }),
        ],
        1,
      );
      renderPage([`/?status=${status}`]);

      await user.click(await screen.findByRole("button", { name: buttonName }));

      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "uid-1", name: "a-container" }),
        expectedAction,
      );
    },
  );
});
