import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import AdminAuthentication from "../Authentication";
import { mockSdkResponse } from "@/tests/sdk";

const mockGetSettings = vi.hoisted(() => vi.fn());
const mockConfigureLocal = vi.hoisted(() => vi.fn());
const mockConfigureSaml = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    getAuthenticationSettings: mockGetSettings,
    configureLocalAuthentication: mockConfigureLocal,
    configureSamlAuthentication: mockConfigureSaml,
  };
});

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
  mockGetSettings.mockReset();
  mockConfigureLocal.mockReset();
  mockConfigureSaml.mockReset();
  mockConfigureLocal.mockResolvedValue(mockSdkResponse(undefined));
  mockConfigureSaml.mockResolvedValue(mockSdkResponse(undefined));
});

describe("AdminAuthentication", () => {
  describe("DS Toggle usage", () => {
    it("renders the local-auth and SAML rows as role='switch' toggles", async () => {
      mockGetSettings.mockResolvedValue(mockSdkResponse(mockSettings()));

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
      mockGetSettings.mockResolvedValue(mockSdkResponse(mockSettings()));

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle local authentication" }),
      );

      expect(mockConfigureLocal).toHaveBeenCalledWith(
        expect.objectContaining({ body: { enable: false } }),
      );
    });

    it("clicking the SAML toggle to turn it off fires configureSamlAuthentication with enable: false", async () => {
      const user = userEvent.setup();
      mockGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );

      renderPage();
      await settlePendingLoad();

      await user.click(
        screen.getByRole("switch", { name: "Toggle SAML authentication" }),
      );

      expect(mockConfigureSaml).toHaveBeenCalledWith(
        expect.objectContaining({
          body: expect.objectContaining({ enable: false }),
        }),
      );
    });

    it("disables the local-auth toggle while togglingLocal is true", async () => {
      const user = userEvent.setup();
      mockGetSettings.mockResolvedValue(mockSdkResponse(mockSettings()));
      let resolveConfigure: (() => void) | undefined;
      mockConfigureLocal.mockReturnValue(
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
      mockGetSettings.mockResolvedValue(
        mockSdkResponse(mockSettings({ samlEnabled: true })),
      );
      let resolveConfigure: (() => void) | undefined;
      mockConfigureSaml.mockReturnValue(
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
