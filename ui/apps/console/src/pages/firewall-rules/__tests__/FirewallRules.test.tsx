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

const capturedDataTableProps: Record<string, unknown>[] = [];
vi.mock("@/components/common/DataTable", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@/components/common/DataTable")>();
  return {
    ...actual,
    default: (props: Record<string, unknown>) => {
      capturedDataTableProps.push({ ...props });
      return actual.default(
        props as unknown as Parameters<typeof actual.default>[0],
      );
    },
  };
});

let lastRulesUrl: URL | null;

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
  capturedDataTableProps.length = 0;
  lastRulesUrl = null;
  useAuthStore.setState({ role: "owner" });
  server.use(
    http.get("*/api/firewall/rules", ({ request }) => {
      lastRulesUrl = new URL(request.url);
      return jsonWithTotal([mockFirewallRule({ priority: 42 })]);
    }),
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
        HttpResponse.json(
          { message: "Permission denied" },
          { status: 403 },
        ),
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
          return HttpResponse.json(
            { message: "Transient" },
            { status: 500 },
          );
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

describe("FirewallRules — URL hydration", () => {
  it("hydrates search from URL on mount", async () => {
    renderPage(["/?search=allow"]);
    await screen.findByText("42");
    expect(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
    ).toHaveValue("allow");
  });

  it("hydrates page from URL and passes it to the API", async () => {
    renderPage(["/?page=3"]);
    await waitFor(() => {
      expect(lastRulesUrl).not.toBeNull();
      expect(lastRulesUrl!.searchParams.get("page")).toBe("3");
    });
  });

  it("passes page=1 when URL has no params", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      expect(lastRulesUrl).not.toBeNull();
      expect(lastRulesUrl!.searchParams.get("page")).toBe("1");
    });
  });

  it("setSearch resets page to 1 in the URL", async () => {
    const user = userEvent.setup();
    renderPage(["/?page=3"]);

    await waitFor(() => {
      expect(lastRulesUrl).not.toBeNull();
      expect(lastRulesUrl!.searchParams.get("page")).toBe("3");
    });

    await user.type(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
      "allow",
    );

    await waitFor(() => {
      expect(lastRulesUrl!.searchParams.get("page")).toBe("1");
    });
  });
});

describe("FirewallRules — pagination suppressed while searching", () => {
  it("passes page/totalPages/onPageChange to DataTable when search is empty", async () => {
    renderPage();
    await screen.findByText("42");
    const last = capturedDataTableProps.at(-1);
    expect(last).toBeDefined();
    expect(last).toHaveProperty("page");
    expect(last).toHaveProperty("totalPages");
    expect(last).toHaveProperty("onPageChange");
  });

  it("omits page/totalPages/onPageChange from DataTable while search is non-empty", async () => {
    const user = userEvent.setup();
    renderPage();

    await screen.findByText("42");

    await user.type(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
      "allow",
    );

    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).toBeDefined();
      expect(last).not.toHaveProperty("page");
      expect(last).not.toHaveProperty("totalPages");
      expect(last).not.toHaveProperty("onPageChange");
    });
  });

  it("re-enables pagination props after search is cleared", async () => {
    const user = userEvent.setup();
    renderPage();

    await screen.findByText("42");

    const searchbox = screen.getByRole("searchbox", {
      name: "Search firewall rules by action, priority, IP, or username",
    });

    await user.type(searchbox, "allow");

    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).not.toHaveProperty("page");
    });

    await user.clear(searchbox);

    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).toBeDefined();
      expect(last).toHaveProperty("page");
      expect(last).toHaveProperty("totalPages");
      expect(last).toHaveProperty("onPageChange");
    });
  });
});
