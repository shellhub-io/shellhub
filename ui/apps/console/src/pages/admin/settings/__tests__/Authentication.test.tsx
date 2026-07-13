import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AdminAuthentication from "../Authentication";

vi.mock("@/client", () => ({
  getAuthenticationSettings: vi.fn(),
  configureLocalAuthentication: vi.fn(),
  configureSamlAuthentication: vi.fn(),
}));

import {
  getAuthenticationSettings,
  configureLocalAuthentication,
  configureSamlAuthentication,
} from "@/client";

const mockedGetSettings = vi.mocked(getAuthenticationSettings);
const mockedConfigureLocal = vi.mocked(configureLocalAuthentication);
const mockedConfigureSaml = vi.mocked(configureSamlAuthentication);

type SdkResponse<T = unknown> = {
  data: T;
  request: Request;
  response: Response;
};

function mockSdkResponse<T>(data: T): SdkResponse<T> {
  return {
    data,
    request: new Request("http://localhost"),
    response: new Response(),
  };
}

function mockSettings({
  localEnabled = true,
  samlEnabled = false,
} = {}) {
  return {
    local: { enabled: localEnabled },
    saml: { enabled: samlEnabled },
  };
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
  mockedGetSettings.mockReset();
  mockedConfigureLocal.mockReset();
  mockedConfigureSaml.mockReset();
  mockedConfigureLocal.mockResolvedValue(mockSdkResponse(undefined));
  mockedConfigureSaml.mockResolvedValue(mockSdkResponse(undefined));
});

describe("AdminAuthentication", () => {
  describe("DS Toggle usage", () => {
    it("renders the local-auth and SAML rows as role='switch' toggles", async () => {
      mockedGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings()),
      );

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
      mockedGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings()),
      );

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      );

      expect(mockedConfigureLocal).toHaveBeenCalledWith(
        expect.objectContaining({ body: { enable: false } }),
      );
    });

    it("clicking the SAML toggle to turn it off fires configureSamlAuthentication with enable: false", async () => {
      const user = userEvent.setup();
      mockedGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      );

      expect(mockedConfigureSaml).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ enable: false }),
        }),
      );
    });

    it("disables the local-auth toggle while togglingLocal is true", async () => {
      const user = userEvent.setup();
      mockedGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings()),
      );
      let resolveConfigure: (() => void) | undefined;
      mockedConfigureLocal.mockReturnValue(
        new Promise((resolve) => {
          resolveConfigure = () => resolve(mockSdkResponse(undefined));
        }) as never,
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

      resolveConfigure?.();
      await waitFor(() => expect(localToggle).not.toBeDisabled());
    });

    it("disables the SAML toggle while togglingSaml is true", async () => {
      const user = userEvent.setup();
      mockedGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );
      let resolveConfigure: (() => void) | undefined;
      mockedConfigureSaml.mockReturnValue(
        new Promise((resolve) => {
          resolveConfigure = () => resolve(mockSdkResponse(undefined));
        }) as never,
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

      resolveConfigure?.();
      await waitFor(() => expect(samlToggle).not.toBeDisabled());
    });
  });
});
