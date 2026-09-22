import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import IdentityAccessGuard from "../IdentityAccessGuard";

function setAccessMode(mode: "legacy" | "identity") {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: false,
            connection_announcement: "",
            ssh_access_mode: mode,
            ssh_legacy_allowed: true,
          },
        }),
      ),
    ),
  );
}

/**
 * Renders the mode the guard read, so a case can wait for the namespace to settle. Asserting on
 * the page alone passes on the first render, before the fetch answers, which hides a guard that
 * redirects on the wrong mode.
 */
function ResolvedMode() {
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");

  return <div>mode: {namespace?.settings?.ssh_access_mode ?? "pending"}</div>;
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
                <ResolvedMode />
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
  it("sends the page to public keys in legacy mode", async () => {
    setAccessMode("legacy");
    renderGuard();

    expect(await screen.findByText("public keys")).toBeInTheDocument();
    expect(screen.queryByText("policies content")).not.toBeInTheDocument();
  });

  it("keeps the page once the namespace settles as identity", async () => {
    setAccessMode("identity");
    renderGuard();

    expect(await screen.findByText("mode: identity")).toBeInTheDocument();
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
