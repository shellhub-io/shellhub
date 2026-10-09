import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { useConnectivityStore } from "@/stores/connectivityStore";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { getConfig, defaultConfig } from "@/env";
import NamespaceGuard from "../NamespaceGuard";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

vi.mock("@/components/layout/SessionMenu", () => ({
  default: () => <div data-testid="session-menu" />,
}));

const mockGetConfig = vi.mocked(getConfig);

afterEach(() => {
  vi.useRealTimers();
});

function serveUser(maxNamespaces: number) {
  server.use(
    http.get("*/api/auth/user", () =>
      HttpResponse.json(
        mockUserAuth({
          email: "alice@example.com",
          max_namespaces: maxNamespaces,
        }),
      ),
    ),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  seedAuthStore({ email: "alice@example.com", tenant: null });
  serveUser(-1);
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json(mockUserAuth()),
    ),
    http.get("*/api/vault", () => new HttpResponse(null, { status: 404 })),
  );
  useConnectivityStore.getState().markUp();
});

function renderGuard(initialPath = "/dashboard") {
  return render(
    <MemoryRouter initialEntries={[initialPath]}>
      <Routes>
        <Route element={<NamespaceGuard />}>
          <Route path="/dashboard" element={<div>dashboard content</div>} />
          <Route path="/account/*" element={<div>account content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("NamespaceGuard", () => {
  it("holds the outlet back behind a spinner while namespaces load", () => {
    server.use(http.get("*/api/namespaces", () => new Promise(() => {})));
    renderGuard();
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
    expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
  });

  it("renders the outlet when namespaces exist", async () => {
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([mockNamespace({ tenant_id: "t1", name: "ns1" })]),
      ),
    );
    renderGuard();
    expect(await screen.findByText("dashboard content")).toBeInTheDocument();
    expect(
      screen.queryByRole("heading", { name: /get your first shell/i }),
    ).not.toBeInTheDocument();
  });

  it("replaces a non-profile route with the first-run screen for getting a namespace", async () => {
    renderGuard("/dashboard");
    expect(
      await screen.findByRole("heading", { name: /join a namespace/i }),
    ).toBeInTheDocument();
    expect(screen.getByTestId("session-menu")).toBeInTheDocument();
    expect(screen.queryByText("dashboard content")).not.toBeInTheDocument();
  });

  it("offers the namespace form, as the start of the trail, to a user who may create one", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    renderGuard("/dashboard");
    expect(
      await screen.findByRole("button", { name: /^create$/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /get your first shell/i }),
    ).toBeInTheDocument();
    expect(screen.getByText(/install the agent/i)).toBeInTheDocument();
  });

  it("offers cloud the namespace form too", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
    renderGuard("/dashboard");
    expect(
      await screen.findByRole("button", { name: /^create$/i }),
    ).toBeInTheDocument();
  });

  it("enters the namespace by itself once an administrator adds the user", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    serveUser(0);
    let added = false;
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal(
          added ? [mockNamespace({ tenant_id: "t1", name: "team" })] : [],
        ),
      ),
    );
    renderGuard("/dashboard");

    await screen.findByText(/waiting to be added/i);
    added = true;
    await vi.advanceTimersByTimeAsync(5000);

    expect(await screen.findByText("dashboard content")).toBeInTheDocument();
  });

  it("offers to try again when getting into the new namespace fails", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    serveUser(0);
    server.use(
      http.get("*/api/namespaces", ({ request }) =>
        jsonWithTotal(
          new URL(request.url).searchParams.get("per_page") === "1"
            ? [mockNamespace({ tenant_id: "t1", name: "team" })]
            : [],
        ),
      ),
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    renderGuard("/dashboard");

    expect(
      await screen.findByText(/getting into the namespace failed/i),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /try again/i }),
    ).toBeInTheDocument();
  });

  it("offers community a retry when getting into the new namespace fails", async () => {
    server.use(
      http.get("*/api/namespaces", ({ request }) =>
        jsonWithTotal(
          new URL(request.url).searchParams.get("per_page") === "1"
            ? [mockNamespace({ tenant_id: "t1", name: "team" })]
            : [],
        ),
      ),
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    renderGuard("/dashboard");

    expect(
      await screen.findByText(/getting into the namespace failed/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /try again/i })).toBeEnabled();
  });

  it("surfaces a refused namespace creation on the name field", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 409 }),
      ),
    );
    const user = userEvent.setup();
    renderGuard("/dashboard");

    await user.type(
      await screen.findByPlaceholderText("my-namespace"),
      "my-ns",
    );
    await user.click(screen.getByRole("button", { name: /^create$/i }));

    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
  });

  it("sends a user who may not create namespaces to an administrator", async () => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
    serveUser(0);
    renderGuard("/dashboard");
    expect(
      await screen.findByText(/ask an administrator to add you/i),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: /join a namespace/i }),
    ).toBeInTheDocument();
    expect(screen.queryByText(/install the agent/i)).not.toBeInTheDocument();
    expect(screen.getByText("alice@example.com")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /^create$/i }),
    ).not.toBeInTheDocument();
  });

  it("gives community the command that adds a member", async () => {
    renderGuard("/dashboard");
    expect(await screen.findByText(/bin\/cli member add/i)).toBeInTheDocument();
  });

  it("lets the account pages through without namespaces", async () => {
    renderGuard("/account/security");
    expect(await screen.findByText("account content")).toBeInTheDocument();
    expect(screen.queryByTestId("session-menu")).not.toBeInTheDocument();
  });
});
