import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import Sidebar from "../Sidebar";

async function renderSidebar(mode: "legacy" | "identity") {
  let served: () => void;
  const namespaceServed = new Promise<void>((resolve) => (served = resolve));
  server.use(
    http.get("*/api/namespaces/:tenant", () => {
      served();
      return HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: false,
            connection_announcement: "",
            ssh_access_mode: mode,
            ssh_legacy_allowed: true,
          },
        }),
      );
    }),
  );
  render(
    <MemoryRouter>
      <Sidebar expanded />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
  await namespaceServed;
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
});

describe("Sidebar", () => {
  it("offers the identity pages only in identity mode", async () => {
    await renderSidebar("identity");

    expect(
      await screen.findByRole("link", { name: /access policies/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /ssh identities/i }),
    ).toBeInTheDocument();
  });

  it("hides the identity pages in legacy mode, where nothing reads them", async () => {
    await renderSidebar("legacy");

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

  it("groups the legacy key pages under SSH", async () => {
    await renderSidebar("legacy");

    const ssh = await screen.findByRole("group", { name: "SSH" });
    expect(
      within(ssh).getByRole("link", { name: /public keys/i }),
    ).toBeInTheDocument();
    expect(screen.queryByText("Security")).not.toBeInTheDocument();
  });

  it("keeps hiding the legacy pages in identity mode", async () => {
    await renderSidebar("identity");

    await screen.findByRole("link", { name: /access policies/i });
    expect(
      screen.queryByRole("link", { name: /public keys/i }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /firewall rules/i }),
    ).not.toBeInTheDocument();
  });
});
