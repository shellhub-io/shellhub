import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useNavigate } from "react-router-dom";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { EXPIRED_JWT, VALID_JWT } from "@/tests/seedAuthStore";
import SetupGuard from "../SetupGuard";

function mockSetup(done: boolean) {
  server.use(http.get("*/info", () => HttpResponse.json({ setup: done })));
}

function holdSetupCheck() {
  const signals: AbortSignal[] = [];
  let releaseResponses: () => void = () => {};
  const released = new Promise<void>((resolve) => {
    releaseResponses = resolve;
  });
  server.use(
    http.get("*/info", async ({ request }) => {
      signals.push(request.signal);
      await released;
      return HttpResponse.json({ setup: false });
    }),
  );
  return {
    signals,
    release: () =>
      act(async () => {
        releaseResponses();
        await released;
      }),
  };
}

function GoToDevices() {
  const navigate = useNavigate();
  return <button type="button" onClick={() => void navigate("/devices")}>go to devices</button>;
}

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <GoToDevices />
      <Routes>
        <Route element={<SetupGuard />}>
          <Route path="/setup" element={<div>setup page</div>} />
          <Route path="/" element={<div>app content</div>} />
          <Route path="/devices" element={<div>devices page</div>} />
        </Route>
        <Route path="/login" element={<div>login page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ token: null });
});

describe("SetupGuard", () => {
  it("cancels an unfinished setup check when unmounted", async () => {
    const check = holdSetupCheck();

    const { unmount } = renderAt("/");
    try {
      await waitFor(() => expect(check.signals).toHaveLength(1));
      unmount();
      await waitFor(() => expect(check.signals[0].aborted).toBe(true));
    } finally {
      await check.release();
    }
  });

  it("ignores a setup check cancelled by navigation", async () => {
    const user = userEvent.setup();
    const check = holdSetupCheck();

    renderAt("/");
    try {
      await waitFor(() => expect(check.signals).toHaveLength(1));
      await user.click(screen.getByRole("button", { name: "go to devices" }));
      await waitFor(() => expect(check.signals[0].aborted).toBe(true));
      await waitFor(() => expect(check.signals).toHaveLength(2));

      expect(screen.getByText("Loading...")).toBeInTheDocument();
      expect(screen.queryByText("devices page")).not.toBeInTheDocument();
    } finally {
      await check.release();
    }

    expect(await screen.findByText("setup page")).toBeInTheDocument();
    expect(screen.queryByText("devices page")).not.toBeInTheDocument();
  });

  it("keeps a just-authenticated user on /setup so the success screen can show", async () => {
    mockSetup(false);
    renderAt("/setup");
    expect(await screen.findByText("setup page")).toBeInTheDocument();

    await act(async () => {
      useAuthStore.setState({ token: VALID_JWT });
    });

    expect(screen.getByText("setup page")).toBeInTheDocument();
    expect(screen.queryByText("login page")).not.toBeInTheDocument();
  });

  it("redirects away from /setup to /login once setup is already done", async () => {
    mockSetup(true);
    renderAt("/setup");
    expect(await screen.findByText("login page")).toBeInTheDocument();
  });

  it("does not bounce an authenticated user on / back to /setup while setup state is stale", async () => {
    mockSetup(false);
    useAuthStore.setState({ token: VALID_JWT });
    renderAt("/");

    expect(await screen.findByText("app content")).toBeInTheDocument();
    expect(screen.queryByText("setup page")).not.toBeInTheDocument();
  });

  it("holds a session whose token expired at the spinner, as it does a signed-out one, until the setup check answers", async () => {
    mockSetup(true);
    useAuthStore.setState({
      token: EXPIRED_JWT,
    });
    renderAt("/");

    expect(screen.getByText("Loading...")).toBeInTheDocument();
    expect(screen.queryByText("app content")).not.toBeInTheDocument();
    expect(await screen.findByText("app content")).toBeInTheDocument();
  });

  it("redirects an unauthenticated visitor to /setup on a fresh install", async () => {
    mockSetup(false);
    renderAt("/");
    expect(await screen.findByText("setup page")).toBeInTheDocument();
  });
});
