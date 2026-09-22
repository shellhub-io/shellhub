import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import Sidebar from "../Sidebar";

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

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
});

function renderSidebar() {
  return render(
    <MemoryRouter>
      <Sidebar expanded pinned={false} onToggle={vi.fn()} />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("Sidebar", () => {
  it("offers the identity pages only in identity mode", async () => {
    setAccessMode("identity");
    renderSidebar();

    expect(
      await screen.findByRole("link", { name: /access policies/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /ssh identities/i }),
    ).toBeInTheDocument();
  });

  it("hides the identity pages in legacy mode, where nothing reads them", async () => {
    setAccessMode("legacy");
    renderSidebar();

    expect(
      await screen.findByRole("link", { name: /public keys/i }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /access policies/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /ssh identities/i }),
    ).not.toBeInTheDocument();
  });

  it("keeps hiding the legacy pages in identity mode", async () => {
    setAccessMode("identity");
    renderSidebar();

    await screen.findByRole("link", { name: /access policies/i });
    expect(
      screen.queryByRole("link", { name: /public keys/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /firewall rules/i }),
    ).not.toBeInTheDocument();
  });
});
