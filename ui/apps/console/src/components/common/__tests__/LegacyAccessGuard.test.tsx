import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockLegacyNamespace, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import LegacyAccessGuard from "../LegacyAccessGuard";

function serveNamespace(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(legacy ? mockLegacyNamespace() : mockNamespace()),
    ),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
});

function renderGuard() {
  return render(
    <MemoryRouter initialEntries={["/sshkeys/public-keys"]}>
      <Routes>
        <Route element={<LegacyAccessGuard />}>
          <Route
            path="/sshkeys/public-keys"
            element={
              <>
                <div>public keys content</div>
                <NamespaceModeProbe />
              </>
            }
          />
        </Route>
        <Route path="/access-policies" element={<div>policies</div>} />
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("LegacyAccessGuard", () => {
  it("sends the page to access policies once the namespace is new", async () => {
    serveNamespace(false);
    renderGuard();

    expect(await screen.findByText("policies")).toBeInTheDocument();
    expect(screen.queryByText("public keys content")).not.toBeInTheDocument();
  });

  it("keeps the page in a legacy namespace", async () => {
    serveNamespace(true);
    renderGuard();

    expect(await screen.findByText("mode: legacy")).toBeInTheDocument();
    expect(screen.getByText("public keys content")).toBeInTheDocument();
    expect(screen.queryByText("policies")).not.toBeInTheDocument();
  });

  it("holds the page while the namespace is still unknown", () => {
    server.use(
      http.get("*/api/namespaces/:tenant", () => new Promise(() => {})),
    );
    renderGuard();

    expect(screen.getByText("public keys content")).toBeInTheDocument();
    expect(screen.queryByText("policies")).not.toBeInTheDocument();
  });
});
