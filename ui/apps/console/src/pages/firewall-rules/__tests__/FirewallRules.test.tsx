import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import FirewallRules from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockFirewallRule } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

vi.mock("../RuleDrawer", () => ({
  default: () => null,
}));

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <FirewallRules />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ role: "owner" });
  server.use(
    http.get("*/api/firewall/rules", () =>
      jsonWithTotal([mockFirewallRule({ priority: 42 })]),
    ),
    http.delete(
      "*/api/firewall/rules/:id",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("FirewallRules — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderPage();
    await screen.findByText("42");
    await user.click(
      screen.getByRole("button", { name: /^delete firewall rule/i }),
    );
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete firewall rule/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    server.use(
      http.delete("*/api/firewall/rules/:id", () =>
        HttpResponse.json({ message: "Permission denied" }, { status: 403 }),
      ),
    );
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("Permission denied")).toBeInTheDocument(),
    );
    expect(dialog).toBeInTheDocument();
  });

  it("shows the status code as fallback when the server returns no message", async () => {
    server.use(
      http.delete("*/api/firewall/rules/:id", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("500")).toBeInTheDocument(),
    );
  });

  it("closes the dialog and does not show an error on successful deletion", async () => {
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: /delete firewall rule/i }),
      ).not.toBeInTheDocument(),
    );
    expect(
      screen.queryByText(/failed to delete firewall rule/i),
    ).not.toBeInTheDocument();
  });

  it("clears any previous error when the dialog is cancelled and reopened", async () => {
    let callCount = 0;
    server.use(
      http.delete("*/api/firewall/rules/:id", () => {
        callCount++;
        if (callCount === 1)
          return HttpResponse.json({ message: "Transient" }, { status: 500 });
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const user = await openDeleteDialog();
    let dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));
    await within(dialog).findByText("Transient");

    await user.click(within(dialog).getByRole("button", { name: /cancel/i }));
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: /delete firewall rule/i }),
      ).not.toBeInTheDocument(),
    );

    await user.click(
      screen.getByRole("button", { name: /^delete firewall rule/i }),
    );
    dialog = await getDialog();
    expect(within(dialog).queryByText("Transient")).not.toBeInTheDocument();
  });
});

describe("FirewallRules — pagination suppressed while searching", () => {
  it("hides the pagination controls while a search is active and restores them when cleared", async () => {
    const user = userEvent.setup();
    server.use(
      http.get("*/api/firewall/rules", () =>
        jsonWithTotal([mockFirewallRule({ priority: 42 })], 30),
      ),
    );
    renderPage();
    await screen.findByText("42");

    expect(
      screen.getByRole("button", { name: /next page/i }),
    ).toBeInTheDocument();

    const searchbox = screen.getByRole("searchbox", {
      name: "Search firewall rules by action, priority, IP, or username",
    });
    await user.type(searchbox, "allow");

    await waitFor(() =>
      expect(
        screen.queryByRole("button", { name: /next page/i }),
      ).not.toBeInTheDocument(),
    );

    await user.clear(searchbox);

    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: /next page/i }),
      ).toBeInTheDocument(),
    );
  });
});
