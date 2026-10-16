import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockLegacyNamespace, mockNamespace } from "@/tests/factories";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import AddDockerConnectorModal from "../AddDockerConnectorModal";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function serveNamespace(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(legacy ? mockLegacyNamespace() : mockNamespace()),
    ),
  );
}

function renderModal() {
  return render(
    <>
      <AddDockerConnectorModal open onClose={() => {}} />
      <NamespaceModeProbe />
    </>,
    {
      wrapper: createTestWrapper({ initialEntries: ["/containers"] }),
    },
  );
}

beforeEach(() => {
  seedAuthStore();
});

describe("AddDockerConnectorModal", () => {
  it("gives a legacy namespace the connector command with its tenant", async () => {
    serveNamespace(true);
    renderModal();

    expect(
      await screen.findByText(/TENANT_ID=tenant-456 .* sh -s connector/),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Open Fleet" }),
    ).not.toBeInTheDocument();
  });

  it("sends a new namespace to Fleet for a provisioning key", async () => {
    serveNamespace(false);
    renderModal();
    await screen.findByText("mode: new");

    expect(
      await screen.findByRole("link", { name: "Open Fleet" }),
    ).toHaveAttribute("href", "/devices/add/fleet");
    expect(screen.queryByText(/TENANT_ID=/)).not.toBeInTheDocument();
  });
});
