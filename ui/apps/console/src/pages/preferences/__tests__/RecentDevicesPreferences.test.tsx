import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { useRecentDevicesStore } from "@/stores/recentDevicesStore";
import RecentDevicesPreferences from "../RecentDevicesPreferences";

const CONNECTED_AT = "2026-09-01T10:00:00Z";

function renderRecent() {
  const user = userEvent.setup();
  render(<RecentDevicesPreferences />, { wrapper: createTestWrapper() });
  return user;
}

beforeEach(() => {
  server.use(
    http.get("*/api/namespaces", () =>
      jsonWithTotal([mockNamespace({ name: "dev", tenant_id: "mine" })]),
    ),
  );
  useRecentDevicesStore.setState({
    byTenant: {
      mine: [
        { uid: "dev-1", name: "web-01", connectedAt: CONNECTED_AT },
        { uid: "dev-2", name: "db-01", connectedAt: CONNECTED_AT },
      ],
      gone: [{ uid: "dev-3", name: "old-box", connectedAt: CONNECTED_AT }],
    },
  });
});

describe("RecentDevicesPreferences", () => {
  it("names a namespace the user left rather than showing its tenant", async () => {
    renderRecent();

    expect(await screen.findByText("dev")).toBeInTheDocument();
    expect(screen.getByText("A namespace you left")).toBeInTheDocument();
    expect(screen.queryByText("gone")).not.toBeInTheDocument();
  });

  it("forgets one device and keeps the others", async () => {
    const user = renderRecent();

    await user.click(
      await screen.findByRole("button", { name: "Forget db-01" }),
    );

    expect(screen.queryByText("db-01")).not.toBeInTheDocument();
    expect(screen.getByText("web-01")).toBeInTheDocument();
  });

  it("clears the whole history", async () => {
    const user = renderRecent();

    await user.click(
      await screen.findByRole("button", { name: "Clear history" }),
    );

    expect(
      screen.getByText("Devices you connect to will appear here."),
    ).toBeInTheDocument();
  });
});
