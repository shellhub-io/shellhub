import { describe, it, expect, beforeEach } from "vitest";
import { act, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { mockLegacyNamespace, mockNamespace } from "@/tests/factories";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import { useAuthStore } from "@/stores/authStore";
import Interactive from "../Interactive";

function serveNamespace(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/tenant-new", () =>
      HttpResponse.json(mockNamespace({ tenant_id: "tenant-new" })),
    ),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(legacy ? mockLegacyNamespace() : mockNamespace()),
    ),
  );
}

function renderPage() {
  return render(
    <ClipboardProvider>
      <Interactive />
      <NamespaceModeProbe />
    </ClipboardProvider>,
    { wrapper: createTestWrapper({ initialEntries: ["/devices/add"] }) },
  );
}

async function pickMethod(label: string, user = userEvent.setup()) {
  await user.click(screen.getByRole("button", { name: /Show all methods/ }));
  await user.click(screen.getByRole("radio", { name: new RegExp(label) }));
}

beforeEach(() => {
  seedAuthStore();
});

describe("Interactive", () => {
  describe("in a new namespace", () => {
    beforeEach(() => serveNamespace(false));

    it("installs with no credential where the installer can pair", async () => {
      renderPage();
      await screen.findByText("mode: new");

      const command = await screen.findByText(/install\.sh/);
      expect(command).not.toHaveTextContent("TENANT_ID");
      expect(command).not.toHaveTextContent("PROVISIONING_KEY");
      expect(
        await screen.findByRole("button", { name: /the code it shows/ }),
      ).toBeInTheDocument();
    });

    it.each(["Snap", "FreeBSD"])(
      "does not offer %s, which enrolls only by tenant",
      async (label) => {
        const user = userEvent.setup();
        renderPage();

        await screen.findByText("mode: new");
        await user.click(
          screen.getByRole("button", { name: /Show all methods/ }),
        );

        expect(
          screen.queryByRole("radio", { name: new RegExp(label) }),
        ).not.toBeInTheDocument();
      },
    );

    it("points a manual method at its guide instead of a command", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText("mode: new");
      await pickMethod("Yocto", user);

      expect(screen.queryByText(/install\.sh/)).not.toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: /Yocto Project guide/ }),
      ).toBeInTheDocument();
    });
  });

  describe("in a legacy namespace", () => {
    beforeEach(() => serveNamespace(true));

    it("puts the tenant on every command and offers no pairing", async () => {
      renderPage();

      expect(await screen.findByText(/TENANT_ID=\S+/)).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /the code it shows/ }),
      ).not.toBeInTheDocument();
      expect(
        await screen.findByRole("link", { name: /pending queue/ }),
      ).toHaveAttribute("href", "/devices/pending");
    });

    it("still offers snap, which enrolls by tenant", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText(/install\.sh/);
      await pickMethod("Snap", user);

      expect(screen.getByText(/install\.sh/)).toHaveTextContent(
        /INSTALL_METHOD=snap TENANT_ID=\S+/,
      );
    });

    it("falls back to Auto Detect when the namespace switches to one without the method", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText("mode: legacy");
      await pickMethod("Snap", user);
      expect(screen.getByText(/install\.sh/)).toHaveTextContent(
        /INSTALL_METHOD=snap/,
      );

      await act(async () => {
        useAuthStore.setState({ tenant: "tenant-new" });
      });

      await screen.findByText("mode: new");
      expect(screen.getByText(/install\.sh/)).not.toHaveTextContent(
        /INSTALL_METHOD=snap/,
      );
      expect(screen.getByRole("radio", { name: /Auto Detect/ })).toBeChecked();
    });

    it("still offers the FreeBSD guide, which enrolls by tenant", async () => {
      const user = userEvent.setup();
      renderPage();

      await screen.findByText(/install\.sh/);
      await pickMethod("FreeBSD", user);

      expect(
        screen.getByRole("link", { name: /FreeBSD guide/ }),
      ).toBeInTheDocument();
    });
  });
});
