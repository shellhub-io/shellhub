import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

await vi.hoisted(async () =>
  (await import("@/tests/viewport")).installViewport(),
);

import { setEdition } from "@/tests/edition";
import SettingsLayout from "../SettingsLayout";

function renderAt(entry: string) {
  render(
    <MemoryRouter initialEntries={[entry]}>
      <Routes>
        <Route path="/settings" element={<SettingsLayout />}>
          <Route path="general" element={<p>General content</p>} />
          <Route path="billing" element={<p>Billing content</p>} />
        </Route>
      </Routes>
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("SettingsLayout", () => {
  beforeEach(() => {
    setEdition("cloud");
    seedAuthStore();
    server.use(
      http.get("*/api/namespaces/:tenant", () =>
        HttpResponse.json(mockNamespace()),
      ),
    );
  });

  it("lands an old billing anchor on the Billing section", () => {
    renderAt("/settings#billing");

    expect(screen.getByText("Billing content")).toBeInTheDocument();
  });

  it("opens General without the anchor", () => {
    renderAt("/settings");

    expect(screen.getByText("General content")).toBeInTheDocument();
  });

  it("lists Billing on cloud", () => {
    renderAt("/settings/general");

    expect(screen.getByRole("link", { name: "Billing" })).toBeInTheDocument();
  });

  it("leaves Billing out elsewhere", () => {
    setEdition("community");
    renderAt("/settings/general");

    expect(screen.getByRole("link", { name: "SSH" })).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "Billing" }),
    ).not.toBeInTheDocument();
  });
});
