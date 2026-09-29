import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import BrowserIdentityPreferences from "../BrowserIdentityPreferences";

function renderIdentity() {
  render(<BrowserIdentityPreferences />, {
    wrapper: createTestWrapper({ initialEntries: ["/preferences"] }),
  });
}

function serveNamespaces(...namespaces: ReturnType<typeof mockNamespace>[]) {
  server.use(http.get("*/api/namespaces", () => jsonWithTotal(namespaces)));
}

beforeEach(() => {
  seedAuthStore();
});

describe("BrowserIdentityPreferences", () => {
  it("says a legacy-mode namespace does not use the key, and offers no way in", async () => {
    serveNamespaces(mockNamespace({ name: "old-way" }));
    renderIdentity();

    expect(
      await screen.findByText("Not used in legacy mode"),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /old-way/ }),
    ).not.toBeInTheDocument();
  });

  it("offers an identity-mode namespace as a way to its SSH identities", async () => {
    serveNamespaces(
      mockNamespace({
        name: "new-way",
        settings: {
          session_record: false,
          connection_announcement: "",
          ssh_access_mode: "identity",
          ssh_legacy_allowed: false,
        },
      }),
    );
    renderIdentity();

    expect(
      await screen.findByRole("button", { name: /new-way/ }),
    ).toHaveTextContent("No key, made on your first connection");
  });

  it("explains the empty list when the user belongs to no namespace", async () => {
    serveNamespaces();
    renderIdentity();

    expect(
      await screen.findByText(
        "Keys appear here once you belong to a namespace.",
      ),
    ).toBeInTheDocument();
  });
});
