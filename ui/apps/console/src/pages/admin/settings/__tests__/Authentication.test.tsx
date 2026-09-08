import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import AdminAuthentication from "../Authentication";

const localSpy = vi.fn();
const samlSpy = vi.fn();

function mockSettings({ localEnabled = true, samlEnabled = false } = {}) {
  return {
    local: { enabled: localEnabled },
    saml: { enabled: samlEnabled },
  };
}

function setSettings(settings: ReturnType<typeof mockSettings>) {
  server.use(
    http.get("*/admin/api/authentication", () =>
      HttpResponse.json(settings),
    ),
  );
}

function renderPage() {
  return render(<AdminAuthentication />);
}

async function settlePendingLoad() {
  await waitFor(() =>
    expect(screen.queryByText(/loading settings/i)).not.toBeInTheDocument(),
  );
}

beforeEach(() => {
  localSpy.mockReset();
  samlSpy.mockReset();
  server.use(
    http.put("*/admin/api/authentication/local", async ({ request }) => {
      localSpy({ body: await request.json() });
      return HttpResponse.json({});
    }),
    http.put("*/admin/api/authentication/saml", async ({ request }) => {
      samlSpy({ body: await request.json() });
      return HttpResponse.json({});
    }),
  );
});

describe("AdminAuthentication", () => {
  describe("DS Toggle usage", () => {
    it("renders the local-auth and SAML rows as role='switch' toggles", async () => {
      setSettings(mockSettings());

      renderPage();
      await settlePendingLoad();

      expect(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      ).toBeInTheDocument();
    });

    it("clicking the local-auth toggle fires configureLocalAuthentication with the flipped value", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings());

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      );

      await waitFor(() => {
        expect(localSpy).toHaveBeenCalledWith(
          expect.objectContaining({ body: { enable: false } }),
        );
      });
    });

    it("clicking the SAML toggle to turn it off fires configureSamlAuthentication with enable: false", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings({ samlEnabled: true }));

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      );

      await waitFor(() => {
        expect(samlSpy).toHaveBeenCalledWith(
          expect.objectContaining({
            body: expect.objectContaining({ enable: false }),
          }),
        );
      });
    });

    it("disables the local-auth toggle while togglingLocal is true", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings());
      let resolveLocal: (() => void) | undefined;
      server.use(
        http.put("*/admin/api/authentication/local", () =>
          new Promise<Response>((resolve) => {
            resolveLocal = () => resolve(HttpResponse.json({}));
          }),
        ),
      );

      renderPage();
      await settlePendingLoad();

      const localToggle = screen.getByRole("switch", {
        name: "Toggle local authentication",
      });
      await user.click(localToggle);

      expect(localToggle).toBeDisabled();
      expect(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      ).not.toBeDisabled();

      resolveLocal?.();
      await waitFor(() => expect(localToggle).not.toBeDisabled());
    });

    it("disables the SAML toggle while togglingSaml is true", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings({ samlEnabled: true }));
      let resolveSaml: (() => void) | undefined;
      server.use(
        http.put("*/admin/api/authentication/saml", () =>
          new Promise<Response>((resolve) => {
            resolveSaml = () => resolve(HttpResponse.json({}));
          }),
        ),
      );

      renderPage();
      await settlePendingLoad();

      const samlToggle = screen.getByRole("switch", {
        name: "Toggle SAML authentication",
      });
      await user.click(samlToggle);

      expect(samlToggle).toBeDisabled();
      expect(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      ).not.toBeDisabled();

      resolveSaml?.();
      await waitFor(() => expect(samlToggle).not.toBeDisabled());
    });
  });
});
