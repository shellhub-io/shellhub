import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import SetupGuard from "../SetupGuard";

vi.mock("@/client", () => ({ getInfo: vi.fn() }));
vi.mock("@/env", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/env")>();
  return {
    ...actual,
    getConfig: vi.fn(() => ({ ...actual.defaultConfig, cloud: false })),
  };
});

import { getInfo } from "@/client";

const mockedGetInfo = vi.mocked(getInfo);

function mockSetup(done: boolean) {
  mockedGetInfo.mockResolvedValue({
    data: { setup: done },
  } as Awaited<ReturnType<typeof getInfo>>);
}

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<SetupGuard />}>
          <Route path="/setup" element={<div>setup page</div>} />
          <Route path="/" element={<div>app content</div>} />
        </Route>
        <Route path="/login" element={<div>login page</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  mockedGetInfo.mockReset();
  useAuthStore.setState({ token: null });
});

describe("SetupGuard", () => {
  it("keeps a just-authenticated user on /setup so the success screen can show", async () => {
    // Reproduce the auto-login sequence: getInfo resolves setup=false first (user is mid-setup),
    // then loginWithToken sets the token synchronously. The guard must NOT bounce to /login.
    mockSetup(false);
    renderAt("/setup");
    expect(await screen.findByText("setup page")).toBeInTheDocument();

    await act(async () => {
      useAuthStore.setState({ token: "issued-token" });
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
    // The post-setup redirect lands on / while getInfo still reports setup=false; the token
    // must keep the user in the app instead of flashing the setup form.
    mockSetup(false);
    useAuthStore.setState({ token: "issued-token" });
    renderAt("/");

    expect(await screen.findByText("app content")).toBeInTheDocument();
    expect(screen.queryByText("setup page")).not.toBeInTheDocument();
  });

  it("redirects an unauthenticated visitor to /setup on a fresh install", async () => {
    mockSetup(false);
    renderAt("/");
    expect(await screen.findByText("setup page")).toBeInTheDocument();
  });
});
