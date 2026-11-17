import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { useQuery } from "@tanstack/react-query";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockInfo } from "@/tests/factories";
import { getInfoOptions } from "@/client";
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
  return render(<AdminAuthentication />, { wrapper: createTestWrapper() });
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

    it("disables the local-auth toggle while the change is pending", async () => {
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

    it("disables the SAML toggle while the change is pending", async () => {
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

  describe("after a change", () => {
    function ServerInfoReader() {
      useQuery(getInfoOptions());
      return null;
    }

    function renderWithServerInfo() {
      let infoRequests = 0;
      server.use(
        http.get("*/info", () => {
          infoRequests += 1;
          return HttpResponse.json(mockInfo());
        }),
      );
      render(
        <>
          <ServerInfoReader />
          <AdminAuthentication />
        </>,
        { wrapper: createTestWrapper() },
      );
      return () => infoRequests;
    }

    it.each([
      {
        method: "local",
        settings: mockSettings(),
        toggle: "Toggle local authentication",
      },
      {
        method: "SAML",
        settings: mockSettings({ samlEnabled: true }),
        toggle: "Toggle SAML authentication",
      },
    ])("loads the server info again after turning $method off", async ({ settings, toggle }) => {
      const user = userEvent.setup();
      setSettings(settings);
      const infoRequests = renderWithServerInfo();
      await settlePendingLoad();
      await waitFor(() => expect(infoRequests()).toBe(1));

      await user.click(screen.getByRole("switch", { name: toggle }));

      await waitFor(() => expect(infoRequests()).toBe(2));
    });

    it("loads the server info again after saving the SAML configuration", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings());
      const infoRequests = renderWithServerInfo();
      await settlePendingLoad();
      await waitFor(() => expect(infoRequests()).toBe(1));

      await user.click(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      );
      await user.click(screen.getByLabelText(/use metadata url/i));
      await user.type(
        screen.getByLabelText(/idp metadata url/i),
        "https://idp.example.com/metadata.xml",
      );
      await user.click(
        screen.getByRole("button", { name: /save configuration/i }),
      );

      await waitFor(() => expect(infoRequests()).toBe(2));
    });

    it("says every method cannot be turned off when the server refuses", async () => {
      const user = userEvent.setup();
      setSettings(mockSettings());
      server.use(
        http.put("*/admin/api/authentication/local", () =>
          HttpResponse.json({}, { status: 400 }),
        ),
      );

      renderPage();
      await settlePendingLoad();
      await user.click(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      );

      expect(
        await screen.findByText("You cannot disable all authentication methods."),
      ).toBeInTheDocument();
    });
  });
});
