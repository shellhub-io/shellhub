import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AdminAuthentication from "../Authentication";
import { mockSdkResponse } from "@/tests/sdk";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getAuthenticationSettings: vi.fn(),
    configureLocalAuthentication: vi.fn(),
    configureSamlAuthentication: vi.fn(),
  }),
);

function mockSettings({ localEnabled = true, samlEnabled = false } = {}) {
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
  sdk.getAuthenticationSettings.mockReset();
  sdk.configureLocalAuthentication.mockReset();
  sdk.configureSamlAuthentication.mockReset();
  sdk.configureLocalAuthentication.mockResolvedValue(mockSdkResponse(undefined));
  sdk.configureSamlAuthentication.mockResolvedValue(mockSdkResponse(undefined));
});

describe("AdminAuthentication", () => {
  describe("DS Toggle usage", () => {
    it("renders the local-auth and SAML rows as role='switch' toggles", async () => {
      sdk.getAuthenticationSettings.mockResolvedValue(mockSdkResponse(mockSettings()));

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
      sdk.getAuthenticationSettings.mockResolvedValue(mockSdkResponse(mockSettings()));

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      );

      expect(sdk.configureLocalAuthentication).toHaveBeenCalledWith(
        expect.objectContaining({ body: { enable: false } }),
      );
    });

    it("clicking the SAML toggle to turn it off fires configureSamlAuthentication with enable: false", async () => {
      const user = userEvent.setup();
      sdk.getAuthenticationSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      );

      expect(sdk.configureSamlAuthentication).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ enable: false }),
        }),
      );
    });

    it("disables the local-auth toggle while togglingLocal is true", async () => {
      const user = userEvent.setup();
      sdk.getAuthenticationSettings.mockResolvedValue(mockSdkResponse(mockSettings()));
      let resolveConfigure: (() => void) | undefined;
      sdk.configureLocalAuthentication.mockReturnValue(
        new Promise((resolve) => {
          resolveConfigure = () => resolve(mockSdkResponse(undefined));
        }),
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
      sdk.getAuthenticationSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );
      let resolveConfigure: (() => void) | undefined;
      sdk.configureSamlAuthentication.mockReturnValue(
        new Promise((resolve) => {
          resolveConfigure = () => resolve(mockSdkResponse(undefined));
        }),
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
