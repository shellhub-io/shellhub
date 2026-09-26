import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLocation } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockLicense } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";
import { getConfig, defaultConfig } from "@/env";
import AdminNavBar from "../AdminNavBar";

const mockGetConfig = vi.mocked(getConfig);

function Pathname() {
  return <output aria-label="Current page">{useLocation().pathname}</output>;
}

function renderBar(compact: boolean, path = "/admin/users") {
  return render(
    <>
      <AdminNavBar compact={compact} />
      <Pathname />
    </>,
    { wrapper: createTestWrapper({ initialEntries: [path] }) },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
  useAuthStore.setState({ isAdmin: true });
  server.use(
    http.get("*/admin/api/license", () =>
      HttpResponse.json(mockLicense({ expired: false })),
    ),
  );
});

describe("AdminNavBar", () => {
  describe("on a wide window", () => {
    it("lays every admin page out as a link", async () => {
      renderBar(false);

      const nav = await screen.findByRole("navigation", {
        name: "Admin navigation",
      });
      expect(nav).toBeInTheDocument();
      expect(
        await screen.findByRole("link", { name: "Namespaces" }),
      ).toBeInTheDocument();
      expect(screen.getByRole("link", { name: "Authentication" })).toBeInTheDocument();
    });

    it("marks the page being viewed", async () => {
      renderBar(false);

      expect(await screen.findByRole("link", { name: "Users" })).toHaveAttribute(
        "aria-current",
        "page",
      );
    });

    it("leaves License out on cloud, which has no license", async () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
      renderBar(false);

      expect(
        await screen.findByRole("link", { name: "Authentication" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("link", { name: "License" }),
      ).not.toBeInTheDocument();
    });

    it("narrows to License until a valid license is known", async () => {
      server.use(
        http.get("*/admin/api/license", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );
      renderBar(false, "/admin/license");

      expect(
        await screen.findByRole("link", { name: "License" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("link", { name: "Dashboard" }),
      ).not.toBeInTheDocument();
    });

    it("greys every page out for someone who is not an instance admin", async () => {
      useAuthStore.setState({ isAdmin: false });
      renderBar(false);

      expect(await screen.findByText("License")).toHaveAttribute(
        "aria-disabled",
        "true",
      );
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    });
  });

  describe("on a narrow window", () => {
    it("names the page being viewed instead of listing them all", async () => {
      renderBar(true);

      expect(
        await screen.findByRole("button", { name: "Admin page: Users" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("link", { name: "Namespaces" }),
      ).not.toBeInTheDocument();
    });

    it("marks the page being viewed in the list", async () => {
      const user = userEvent.setup();
      renderBar(true);

      await user.click(
        await screen.findByRole("button", { name: "Admin page: Users" }),
      );

      expect(screen.getByRole("menuitem", { name: "Users" })).toHaveAttribute(
        "aria-current",
        "page",
      );
      expect(
        screen.getByRole("menuitem", { name: "Devices" }),
      ).not.toHaveAttribute("aria-current");
    });

    it("opens onto the other pages and goes to the one picked", async () => {
      const user = userEvent.setup();
      renderBar(true);

      await user.click(
        await screen.findByRole("button", { name: "Admin page: Users" }),
      );
      await user.click(screen.getByRole("menuitem", { name: "Devices" }));

      expect(screen.getByLabelText("Current page")).toHaveTextContent(
        "/admin/devices",
      );
    });
  });
});
