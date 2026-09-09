import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockStats } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";

vi.mock("@/utils/welcomeState", () => ({
  hasSeenWelcome: vi.fn(),
  markWelcomeSeen: vi.fn(),
}));

vi.mock("../WelcomeWizard", () => ({
  default: ({
    open,
    onClose,
    onDismiss,
  }: {
    open: boolean;
    onClose: () => void;
    onDismiss: () => void;
  }) =>
    open ? (
      <>
        <button
          type="button"
          aria-label="Close wizard"
          data-testid="welcome-wizard"
          onClick={onClose}
        />
        <button
          type="button"
          aria-label="Dismiss wizard"
          data-testid="wizard-dismiss"
          onClick={onDismiss}
        />
      </>
    ) : null,
}));

import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizardTrigger from "../WelcomeWizardTrigger";

const mockHasSeenWelcome = vi.mocked(hasSeenWelcome);
const mockMarkWelcomeSeen = vi.mocked(markWelcomeSeen);

function renderTrigger() {
  return render(<WelcomeWizardTrigger />, { wrapper: createTestWrapper() });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(http.get("*/api/stats", () => HttpResponse.json(mockStats())));
  mockHasSeenWelcome.mockReturnValue(false);
});

describe("WelcomeWizardTrigger", () => {
  describe("when tenant has already seen welcome", () => {
    it("renders nothing", async () => {
      mockHasSeenWelcome.mockReturnValue(true);
      server.use(http.get("*/api/stats", () => HttpResponse.json(mockStats())));
      renderTrigger();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
    });
  });

  describe("when tenant has not seen welcome and has zero devices", () => {
    beforeEach(() => {
      server.use(http.get("*/api/stats", () => HttpResponse.json(mockStats())));
    });

    it("shows the wizard", async () => {
      renderTrigger();
      expect(await screen.findByTestId("welcome-wizard")).toBeInTheDocument();
    });

    it("hides the wizard when closed without marking it seen", async () => {
      renderTrigger();
      (await screen.findByTestId("welcome-wizard")).click();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });

    it("calls markWelcomeSeen with the tenant id when dismissed for good", async () => {
      renderTrigger();
      (await screen.findByTestId("wizard-dismiss")).click();
      await waitFor(() => {
        expect(mockMarkWelcomeSeen).toHaveBeenCalledWith("tenant-456");
      });
    });

    it("refetches stats when wizard is closed", async () => {
      let callCount = 0;
      server.use(
        http.get("*/api/stats", () => {
          callCount++;
          return HttpResponse.json(mockStats());
        }),
      );
      renderTrigger();
      const before = callCount;
      (await screen.findByTestId("welcome-wizard")).click();
      await waitFor(() => {
        expect(callCount).toBeGreaterThan(before);
      });
    });
  });

  describe("when tenant has devices", () => {
    it("neither shows the wizard nor marks it seen when devices are registered", async () => {
      server.use(
        http.get("*/api/stats", () =>
          HttpResponse.json(mockStats({ registered_devices: 1 })),
        ),
      );
      renderTrigger();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
      expect(mockMarkWelcomeSeen).not.toHaveBeenCalled();
    });

    it.each(["pending_devices", "rejected_devices"] as const)(
      "still shows the wizard when a device is only counted in %s",
      async (field) => {
        server.use(
          http.get("*/api/stats", () =>
            HttpResponse.json(mockStats({ [field]: 1 })),
          ),
        );
        renderTrigger();
        expect(await screen.findByTestId("welcome-wizard")).toBeInTheDocument();
      },
    );
  });

  describe("eligibility is decided once, at page load", () => {
    it("does not reopen when the last device is deleted mid-session", async () => {
      server.use(
        http.get("*/api/stats", () =>
          HttpResponse.json(mockStats({ registered_devices: 1 })),
        ),
      );
      const { rerender } = renderTrigger();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });

      server.use(http.get("*/api/stats", () => HttpResponse.json(mockStats())));
      rerender(<WelcomeWizardTrigger />);
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when stats are loading", () => {
    it("does not show the wizard", () => {
      server.use(http.get("*/api/stats", () => new Promise(() => {})));
      renderTrigger();
      expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
    });
  });

  describe("when tenant is null", () => {
    it("does not show the wizard", async () => {
      seedAuthStore({ tenant: null });
      server.use(http.get("*/api/stats", () => HttpResponse.json(mockStats())));
      renderTrigger();
      await waitFor(() => {
        expect(screen.queryByTestId("welcome-wizard")).not.toBeInTheDocument();
      });
    });
  });
});
