import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { decodeB64url } from "@/tests/decodeB64url";
import { setEdition } from "@/tests/edition";
import { mockAdminUser, mockLicense, mockNamespace } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";
import type { GetLicenseResponse, UserAdminResponse } from "@/client";
import AdminInstance from "../Instance";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

const maria = mockAdminUser({
  id: "user-maria",
  name: "Maria",
  email: "maria@acme.io",
  username: "maria",
});

const stats = {
  registered_users: 42,
  registered_devices: 9,
  online_devices: 3,
  pending_devices: 2,
  rejected_devices: 1,
  active_sessions: 4,
};

type FilterProperty = {
  type: "property";
  params: { name: string; operator: string; value: unknown };
};

const SUBSET_PROPERTIES: Record<string, FilterProperty["params"]> = {
  awaiting_approval: {
    name: "awaiting_approval",
    operator: "bool",
    value: true,
  },
  admin: { name: "admin", operator: "bool", value: true },
  not_confirmed: { name: "status", operator: "eq", value: "not-confirmed" },
};

function requestedSubset(request: Request): string {
  const raw = new URL(request.url).searchParams.get("filter");
  if (!raw) return "";
  const nodes = decodeB64url(raw) as { type: string; params: unknown }[];
  const found = Object.entries(SUBSET_PROPERTIES).find(([, params]) =>
    nodes.some(
      (node) =>
        node.type === "property" &&
        JSON.stringify(node.params) === JSON.stringify(params),
    ),
  );
  return found?.[0] ?? "";
}

function serveUsers({
  pending = [] as UserAdminResponse[],
  admins = 1,
  unconfirmed = 0,
} = {}) {
  const state = { pending };
  server.use(
    http.get("*/admin/api/users", ({ request }) => {
      const subset = requestedSubset(request);
      if (subset === "awaiting_approval") return jsonWithTotal(state.pending);
      if (subset === "admin") return jsonWithTotal([], admins);
      if (subset === "not_confirmed") return jsonWithTotal([], unconfirmed);
      return jsonWithTotal([]);
    }),
  );
  return state;
}

function serveLicense(overrides: Partial<GetLicenseResponse> = {}) {
  server.use(
    http.get("*/admin/api/license", () =>
      HttpResponse.json(mockLicense(overrides)),
    ),
  );
}

function withDeviceLimit(devices: number): Partial<GetLicenseResponse> {
  return { features: { ...mockLicense().features, devices } };
}

function serveStats(registeredDevices: number) {
  server.use(
    http.get("*/admin/api/stats", () =>
      HttpResponse.json({ ...stats, registered_devices: registeredDevices }),
    ),
  );
}

function failStats() {
  server.use(
    http.get("*/admin/api/stats", () => HttpResponse.json({}, { status: 500 })),
  );
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminInstance />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  setEdition("enterprise");
  serveUsers();
  serveLicense();
  server.use(
    http.get("*/admin/api/stats", () => HttpResponse.json(stats)),
    http.get("*/info", () =>
      HttpResponse.json({
        version: "v0.27.0",
        endpoints: { ssh: "shellhub.example:22", api: "shellhub.example:443" },
        setup: true,
        authentication: { local: true, saml: true },
      }),
    ),
    http.get("*/admin/api/namespaces", () =>
      jsonWithTotal([mockNamespace({ name: "acme", tenant_id: "t-acme" })], 7),
    ),
    http.get("*/admin/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          name: "acme",
          tenant_id: "t-acme",
          members: [
            {
              id: "user-maria",
              email: "maria@acme.io",
              role: "operator",
              awaiting_approval: true,
            },
          ],
        }),
      ),
    ),
  );
});

describe("AdminInstance", () => {
  describe("enterprise", () => {
    describe("member requests", () => {
      it("lists each member request with the namespace they are joining", async () => {
        serveUsers({ pending: [maria] });
        renderPage();

        const requests = await screen.findByRole("list", {
          name: "New members awaiting approval",
        });
        expect(within(requests).getByText("Maria")).toBeInTheDocument();
        expect(
          await within(requests).findByText(/joining acme/),
        ).toBeInTheDocument();
        expect(
          screen.getByRole("link", { name: /open in the users list/i }),
        ).toHaveAttribute("href", "/admin/users?subset=awaiting_approval");
      });

      it("approves a member request and drops it from the list", async () => {
        const users = serveUsers({ pending: [maria] });
        const approved = vi.fn();
        server.use(
          http.post("*/admin/api/users/:id/approve", ({ params }) => {
            approved(params.id);
            users.pending = [];
            return new HttpResponse(null, { status: 200 });
          }),
        );
        renderPage();

        await userEvent.click(
          await screen.findByRole("button", {
            name: "Approve account for maria@acme.io",
          }),
        );
        const dialog = await screen.findByRole("dialog", {
          name: "Approve user",
        });
        await userEvent.click(
          within(dialog).getByRole("button", { name: "Approve" }),
        );

        await waitFor(() =>
          expect(approved).toHaveBeenCalledWith("user-maria"),
        );
        expect(
          await screen.findByText("Nothing needs your attention."),
        ).toBeInTheDocument();
        expect(screen.queryByText("Maria")).not.toBeInTheDocument();
      });

      it("rejects a member request by deleting the pending user and drops it from the list", async () => {
        const users = serveUsers({ pending: [maria] });
        const deleted = vi.fn();
        server.use(
          http.delete("*/admin/api/users/:id", ({ params }) => {
            deleted(params.id);
            users.pending = [];
            return new HttpResponse(null, { status: 200 });
          }),
        );
        renderPage();

        await userEvent.click(
          await screen.findByRole("button", {
            name: "Reject account for maria@acme.io",
          }),
        );
        const dialog = await screen.findByRole("dialog", {
          name: "Reject user",
        });
        await userEvent.click(
          within(dialog).getByRole("button", { name: "Reject" }),
        );

        await waitFor(() => expect(deleted).toHaveBeenCalledWith("user-maria"));
        await waitFor(() =>
          expect(screen.queryByText("Maria")).not.toBeInTheDocument(),
        );
      });

      it("keeps the rest of the page when member requests fail to load", async () => {
        server.use(
          http.get("*/admin/api/users", ({ request }) =>
            requestedSubset(request) === "awaiting_approval"
              ? HttpResponse.json({}, { status: 500 })
              : jsonWithTotal([], 1),
          ),
        );
        renderPage();

        expect(
          await screen.findByText("Couldn't load member requests."),
        ).toBeInTheDocument();
        expect(
          screen.queryByText("Nothing needs your attention."),
        ).not.toBeInTheDocument();
        expect(
          screen.getByRole("region", { name: "License" }),
        ).toBeInTheDocument();
      });
    });

    describe("needs attention", () => {
      it("says nothing needs attention once requests and the license have answered", async () => {
        renderPage();

        expect(
          await screen.findByText("Nothing needs your attention."),
        ).toBeInTheDocument();
      });

      it("reports a failed license check instead of saying nothing needs attention", async () => {
        server.use(
          http.get("*/admin/api/license", () =>
            HttpResponse.json({}, { status: 500 }),
          ),
        );
        renderPage();

        const panel = screen.getByRole("region", {
          name: "Needs your attention",
        });
        expect(
          await within(panel).findByText(
            "Couldn't check the license and its device limit.",
          ),
        ).toBeInTheDocument();
        expect(
          within(panel).queryByText("Nothing needs your attention."),
        ).not.toBeInTheDocument();
      });

      it("reports a failed device count in both panels and dashes the users count", async () => {
        failStats();
        renderPage();

        expect(
          await screen.findByText("Couldn't load the accepted device count."),
        ).toBeInTheDocument();
        expect(
          screen.getByText("Couldn't check the license and its device limit."),
        ).toBeInTheDocument();
        expect(
          screen.queryByText("Nothing needs your attention."),
        ).not.toBeInTheDocument();
        const access = screen.getByRole("region", { name: "Access" });
        expect(
          within(access).getByRole("link", { name: /^users\s*—/i }),
        ).toBeInTheDocument();
      });

      it.each([
        {
          case: "a license about to expire",
          license: { about_to_expire: true },
          used: 1,
          notice: "Your license is about to expire!",
        },
        {
          case: "a license in its grace period",
          license: { expired: true, grace_period: true },
          used: 1,
          notice:
            "Your license has expired, but you are still within the grace period.",
        },
        {
          case: "a license that has not started",
          license: { starts_at: Date.UTC(2100, 0, 15, 12) / 1000 },
          used: 1,
          notice:
            "This license starts on Jan 15, 2100. Devices can't be accepted until then.",
        },
        {
          case: "accepted devices at 90% of the limit",
          license: withDeviceLimit(10),
          used: 9,
          notice: "9 of 10 licensed devices are in use.",
        },
        {
          case: "accepted devices at the limit",
          license: withDeviceLimit(9),
          used: 9,
          notice:
            "All 9 licensed devices are in use. New devices can't be accepted.",
        },
        {
          case: "accepted devices past the limit",
          license: withDeviceLimit(6),
          used: 9,
          notice:
            "9 devices are accepted against a limit of 6. Connections to devices are refused until the count is back within the license.",
        },
      ])("warns about $case", async ({ license, used, notice }) => {
        serveLicense(license);
        serveStats(used);
        renderPage();

        expect(await screen.findByText(notice)).toBeInTheDocument();
        expect(
          screen.queryByText("Nothing needs your attention."),
        ).not.toBeInTheDocument();
      });
    });

    describe("license", () => {
      it("shows accepted devices against the licensed limit", async () => {
        serveLicense(withDeviceLimit(9));
        serveStats(9);
        renderPage();

        const bar = await screen.findByRole("progressbar", {
          name: "Licensed devices in use",
        });
        expect(bar).toHaveAttribute("aria-valuenow", "9");
        expect(bar).toHaveAttribute("aria-valuemax", "9");
      });

      it("keeps the bar's value in range when accepted devices exceed the limit", async () => {
        serveLicense(withDeviceLimit(6));
        serveStats(9);
        renderPage();

        const bar = await screen.findByRole("progressbar", {
          name: "Licensed devices in use",
        });
        expect(bar).toHaveAttribute("aria-valuenow", "6");
        expect(bar).toHaveAttribute("aria-valuemax", "6");
        expect(bar).toHaveAttribute("aria-valuetext", "9 of 6 devices");
      });

      it("reads an unlimited, non-expiring license as such", async () => {
        serveLicense({ expires_at: 0 });
        renderPage();

        expect(
          await screen.findByText("accepted, unlimited"),
        ).toBeInTheDocument();
        const license = screen.getByRole("region", { name: "License" });
        expect(within(license).getByText("Never")).toBeInTheDocument();
        expect(screen.queryByRole("progressbar")).not.toBeInTheDocument();
      });

      it("names who the license was issued to", async () => {
        serveLicense({
          customer: {
            id: "cus-1",
            name: "Ana",
            email: "a@b.c",
            company: "Acme",
          },
        });
        renderPage();

        const license = screen.getByRole("region", { name: "License" });
        expect(await within(license).findByText("Acme")).toBeInTheDocument();
      });

      it("says the license is loading until it answers", () => {
        server.use(
          http.get("*/admin/api/license", () => new Promise(() => {})),
        );
        renderPage();

        const license = screen.getByRole("region", { name: "License" });
        expect(within(license).getByRole("status")).toHaveTextContent(
          "Loading the license…",
        );
      });

      it("says in the License panel when the license failed to load", async () => {
        server.use(
          http.get("*/admin/api/license", () =>
            HttpResponse.json({}, { status: 500 }),
          ),
        );
        renderPage();

        const license = screen.getByRole("region", { name: "License" });
        expect(
          await within(license).findByText("Couldn't load the license."),
        ).toBeInTheDocument();
      });
    });

    describe("access", () => {
      it("counts users, namespaces and instance admins, linking each to its list", async () => {
        serveUsers({ admins: 2 });
        renderPage();

        const access = screen.getByRole("region", { name: "Access" });
        await waitFor(() =>
          expect(
            within(access).getByRole("link", { name: /^users\s*42/i }),
          ).toHaveAttribute("href", "/admin/users"),
        );
        expect(
          within(access).getByRole("link", { name: /^namespaces\s*7/i }),
        ).toHaveAttribute("href", "/admin/namespaces");
        expect(
          within(access).getByRole("link", { name: /^instance admins\s*2/i }),
        ).toHaveAttribute("href", "/admin/users?subset=admin");
      });

      it("lists the enabled sign-in methods", async () => {
        renderPage();

        const access = screen.getByRole("region", { name: "Access" });
        expect(
          await within(access).findByText("Local · SAML"),
        ).toBeInTheDocument();
      });
    });

    describe("server", () => {
      it("shows the server version and endpoints", async () => {
        renderPage();

        const serverPanel = screen.getByRole("region", { name: "Server" });
        expect(
          await within(serverPanel).findByText("v0.27.0"),
        ).toBeInTheDocument();
        expect(
          within(serverPanel).getByText("shellhub.example:22"),
        ).toBeInTheDocument();
      });

      it("shows server information as loading, not failed, until it answers", () => {
        server.use(http.get("*/info", () => new Promise(() => {})));
        renderPage();

        const serverPanel = screen.getByRole("region", { name: "Server" });
        expect(within(serverPanel).getAllByText("…")).toHaveLength(3);
        expect(within(serverPanel).queryByText("—")).not.toBeInTheDocument();
      });
    });
  });

  describe("cloud", () => {
    beforeEach(() => {
      setEdition("cloud");
    });

    it("leads with the instance figures and shows no people or license", async () => {
      serveUsers({ unconfirmed: 5 });
      renderPage();

      const numbers = screen.getByRole("region", { name: "Instance numbers" });
      expect(await within(numbers).findByText("42")).toBeInTheDocument();
      expect(within(numbers).getByText("4")).toBeInTheDocument();
      expect(
        await within(numbers).findByText("5 haven't confirmed their email"),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("region", { name: "Needs your attention" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("region", { name: "License" }),
      ).not.toBeInTheDocument();
    });

    it("dashes the figures it could not load", async () => {
      failStats();
      renderPage();

      const numbers = screen.getByRole("region", { name: "Instance numbers" });
      expect(
        await within(numbers).findByText(
          "Couldn't load some of these numbers.",
        ),
      ).toBeInTheDocument();
      expect(within(numbers).getAllByText("—")).toHaveLength(4);
    });
  });
});
