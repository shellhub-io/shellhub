import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import IdentityAccessGuard from "../IdentityAccessGuard";

function serveNamespace(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: false,
            connection_announcement: "",
            legacy,
            require_verified_device_key: false,
          },
        }),
      ),
    ),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
});

function renderGuard() {
  return render(
    <MemoryRouter initialEntries={["/access-policies"]}>
      <Routes>
        <Route element={<IdentityAccessGuard />}>
          <Route
            path="/access-policies"
            element={
              <>
                <div>policies content</div>
                <NamespaceModeProbe />
              </>
            }
          />
        </Route>
        <Route path="/sshkeys/public-keys" element={<div>public keys</div>} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("IdentityAccessGuard", () => {
  it("sends the page to public keys in a legacy namespace", async () => {
    serveNamespace(true);
    renderGuard();

    expect(await screen.findByText("public keys")).toBeInTheDocument();
    expect(screen.queryByText("policies content")).not.toBeInTheDocument();
  });

  it("keeps the page once the namespace settles as identity", async () => {
    serveNamespace(false);
    renderGuard();

    expect(await screen.findByText("mode: new")).toBeInTheDocument();
    expect(screen.getByText("policies content")).toBeInTheDocument();
    expect(screen.queryByText("public keys")).not.toBeInTheDocument();
  });

  it("holds the page while the namespace is still unknown", () => {
    server.use(
      http.get("*/api/namespaces/:tenant", () => new Promise(() => {})),
    );
    renderGuard();

    expect(screen.getByText("policies content")).toBeInTheDocument();
    expect(screen.queryByText("public keys")).not.toBeInTheDocument();
  });
});
