import { beforeEach, describe, expect, it } from "vitest";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Link, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { mockNamespace } from "@/tests/factories";
import { connectDialogHandlers } from "@/tests/handlers";
import { useTerminalStore } from "@/stores/terminalStore";
import { useWorkspaceTabsStore } from "@/stores/workspaceTabsStore";
import TerminalManager from "../TerminalManager";

function renderManager(queryClient = createTestQueryClient()) {
  return render(
    <>
      <Routes>
        <Route path="/devices" element={<Link to="/sessions">Sessions</Link>} />
        <Route path="/sessions" element={<h1>Sessions page</h1>} />
        <Route path="/dashboard" element={<h1>Dashboard page</h1>} />
        <Route path="*" element={null} />
      </Routes>
      <TerminalManager />
    </>,
    {
      wrapper: createTestWrapper({
        initialEntries: ["/devices"],
        queryClient,
      }),
    },
  );
}

beforeEach(() => {
  seedAuthStore({ tenant: "tenant-home" });
  useWorkspaceTabsStore.setState({ tabs: [], failures: {} });
  useTerminalStore.setState({ sessions: [], reconnectTarget: null });
  server.use(
    http.get("*/api/namespaces", () =>
      jsonWithTotal([
        mockNamespace({ tenant_id: "tenant-home", name: "home" }),
        mockNamespace({ tenant_id: "tenant-other", name: "other" }),
      ]),
    ),
    http.get("*/api/namespaces/:tenant", ({ params }) =>
      HttpResponse.json(mockNamespace({ tenant_id: String(params.tenant) })),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: VALID_JWT, role: "owner" }),
    ),
    http.post(
      "*/ws/ssh/session",
      () => new HttpResponse(null, { status: 500 }),
    ),
    ...connectDialogHandlers,
  );
});

describe("TerminalManager connect dialog", () => {
  it("closes once the user leaves the page it opened on", async () => {
    const user = userEvent.setup();
    renderManager();

    act(() => useTerminalStore.getState().connect("dev-1", "edge-1"));
    await screen.findByRole("dialog", { name: "Connect" });
    await user.click(screen.getByRole("link", { name: "Sessions" }));

    await screen.findByRole("heading", { name: "Sessions page" });
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: "Connect" }),
      ).not.toBeInTheDocument(),
    );
  });

  it("opens in the device's namespace after entering it", async () => {
    const queryClient = createTestQueryClient();
    renderManager(queryClient);
    await waitFor(() => expect(queryClient.isFetching()).toBe(0));

    act(() => {
      useTerminalStore.getState().open({
        deviceUid: "dev-1",
        deviceName: "edge-1",
        username: "root",
        password: "",
        tenant: "tenant-other",
      });
    });
    const [session] = useTerminalStore.getState().sessions;
    act(() => useTerminalStore.getState().closeAndReconnect(session.id));

    await screen.findByRole("heading", { name: "Dashboard page" });
    const dialog = await screen.findByRole("dialog", { name: "Connect" });
    expect(within(dialog).getByTitle(/^other\.edge-1@/)).toBeInTheDocument();
  });
});
