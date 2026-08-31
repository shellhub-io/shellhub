import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, act } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import SetupGuard from "../SetupGuard";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getInfo: vi.fn(),
  }),
);

function mockSetup(done: boolean) {
  sdk.getInfo.mockResolvedValue({
    data: { setup: done },
  });
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
  sdk.getInfo.mockReset();
  useAuthStore.setState({ token: null });
});

describe("SetupGuard", () => {
  it("keeps a just-authenticated user on /setup so the success screen can show", async () => {
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
