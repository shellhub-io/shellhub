import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse } from "@/tests/sdk";
import { mockStats, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import type { Edition } from "@/env";
import { getConfig } from "@/env";
import DeviceChooserTrigger from "../DeviceChooserTrigger";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getStatusDevices: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

vi.mock("../DeviceChooserDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open
      ? React.createElement(
          "div",
          { "data-testid": "device-chooser-dialog" },
          React.createElement(
            "button",
            { type: "button", onClick: onClose },
            "Dismiss",
          ),
        )
      : null,
}));

const mockGetConfig = vi.mocked(getConfig);

function setupMocks({
  edition = "cloud",
  role = "owner",
  billingActive = false,
  registeredDevices = 4,
  nsLoading = false,
  statsLoading = false,
  namespaceNull = false,
}: {
  edition?: Edition;
  role?: "owner" | "observer";
  billingActive?: boolean;
  registeredDevices?: number;
  nsLoading?: boolean;
  statsLoading?: boolean;
  namespaceNull?: boolean;
} = {}) {
  mockGetConfig.mockReturnValue({ edition } as ReturnType<typeof getConfig>);
  seedAuthStore({ role, isAdmin: role === "owner" });

  if (statsLoading) {
    sdk.getStatusDevices.mockReturnValue(new Promise(() => {}));
  } else {
    sdk.getStatusDevices.mockResolvedValue(
      mockSdkResponse(mockStats({ registered_devices: registeredDevices })),
    );
  }

  if (nsLoading) {
    sdk.getNamespace.mockReturnValue(new Promise(() => {}));
  } else if (namespaceNull) {
    sdk.getNamespace.mockResolvedValue(mockSdkResponse(null));
  } else {
    sdk.getNamespace.mockResolvedValue(
      mockSdkResponse(
        mockNamespace({
          billing: billingActive
            ? {
                active: true,
                status: "active" as const,
                customer_id: "cus_123",
                subscription_id: "sub_123",
                current_period_end: 0,
                created_at: "2024-01-01T00:00:00Z",
                updated_at: "2024-01-01T00:00:00Z",
              }
            : null,
        }),
      ),
    );
  }
}

beforeEach(() => {
  vi.clearAllMocks();
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  setupMocks();
});

function renderTrigger() {
  return render(<DeviceChooserTrigger />, { wrapper: createTestWrapper() });
}

describe("DeviceChooserTrigger", () => {
  describe("when edition isn't cloud", () => {
    it.each(["community", "enterprise"] as const)(
      "renders nothing without mounting the inner component for edition=%s",
      (edition) => {
        setupMocks({ edition });
        renderTrigger();
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument();
      },
    );
  });

  describe("when edition=cloud but user lacks device:choose permission", () => {
    it("renders nothing", async () => {
      setupMocks({ role: "observer" });
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when edition=cloud, owner, billing is active", () => {
    it("renders nothing", async () => {
      setupMocks({ billingActive: true });
      renderTrigger();
      await waitFor(() => expect(sdk.getNamespace).toHaveBeenCalled());
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when edition=cloud, owner, billing inactive, registered_devices=3 (boundary)", () => {
    it("renders nothing — limit is strictly greater than 3", async () => {
      setupMocks({ registeredDevices: 3 });
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when all conditions are met (cloud, owner, billing inactive, >3 devices)", () => {
    it("auto-opens the dialog on mount", async () => {
      renderTrigger();
      expect(
        await screen.findByTestId("device-chooser-dialog"),
      ).toBeInTheDocument();
    });
  });

  describe("when namespace is still loading", () => {
    it("renders nothing even if stats indicate over-limit", () => {
      setupMocks({ nsLoading: true, registeredDevices: 4 });
      renderTrigger();
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when stats are still loading", () => {
    it("renders nothing", () => {
      setupMocks({ statsLoading: true });
      renderTrigger();
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when namespace returned is undefined", () => {
    it("renders nothing — billing is unknown until namespace resolves", async () => {
      setupMocks({ namespaceNull: true, registeredDevices: 4 });
      renderTrigger();
      await waitFor(() => expect(sdk.getNamespace).toHaveBeenCalled());
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when the stats query settled with an error", () => {
    it("renders nothing — overLimit cannot be evaluated without stats", async () => {
      sdk.getStatusDevices.mockRejectedValue({ status: 500 });
      renderTrigger();
      await waitFor(() => expect(sdk.getStatusDevices).toHaveBeenCalled());
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });
  });

  describe("dismissal", () => {
    it("hides the dialog after dismissal", async () => {
      const user = userEvent.setup();
      renderTrigger();
      expect(
        await screen.findByTestId("device-chooser-dialog"),
      ).toBeInTheDocument();
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );
    });

    it("does not re-open the dialog after dismissal within the same mount", async () => {
      const user = userEvent.setup();
      const { rerender } = renderTrigger();
      expect(
        await screen.findByTestId("device-chooser-dialog"),
      ).toBeInTheDocument();
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );

      rerender(<DeviceChooserTrigger />);
      expect(
        screen.queryByTestId("device-chooser-dialog"),
      ).not.toBeInTheDocument();
    });

    it("reopens the dialog on a fresh remount after conditions are still true", async () => {
      const user = userEvent.setup();
      const wrapper = createTestWrapper();

      render(<DeviceChooserTrigger />, { wrapper });
      expect(
        await screen.findByTestId("device-chooser-dialog"),
      ).toBeInTheDocument();
      await user.click(screen.getByRole("button", { name: /dismiss/i }));
      await waitFor(() =>
        expect(
          screen.queryByTestId("device-chooser-dialog"),
        ).not.toBeInTheDocument(),
      );

      cleanup();

      render(<DeviceChooserTrigger />, { wrapper });
      expect(
        await screen.findByTestId("device-chooser-dialog"),
      ).toBeInTheDocument();
    });
  });
});
