import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Announcement } from "@/client";

// ── Dependency mocks ──────────────────────────────────────────────────────────

vi.mock("@/env", () => ({ getConfig: vi.fn() }));

vi.mock("@/hooks/useLatestAnnouncement", () => ({
  useLatestAnnouncement: vi.fn(),
}));

// Replace AnnouncementModal with a simple stub that exposes the open state
// and wires up onClose — we don't need to render the full modal here.
vi.mock("../AnnouncementModal", () => ({
  default: ({
    open,
    onClose,
    announcement,
  }: {
    open: boolean;
    onClose: () => void;
    announcement: Announcement;
  }) =>
    open ? (
      <div data-testid="announcement-modal">
        <span data-testid="modal-title">{announcement.title}</span>
        <button onClick={onClose}>Close</button>
      </div>
    ) : null,
}));

import { getConfig } from "@/env";
import { useLatestAnnouncement } from "@/hooks/useLatestAnnouncement";
import AnnouncementModalTrigger from "../AnnouncementModalTrigger";

const mockGetConfig = vi.mocked(getConfig);
const mockUseLatestAnnouncement = vi.mocked(useLatestAnnouncement);

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    uuid: "ann-uuid-1",
    title: "Test Announcement",
    content: "## Content",
    date: "2024-06-01T00:00:00Z",
    ...overrides,
  };
}

function computeHash(announcement: Announcement): string {
  return btoa(JSON.stringify(announcement));
}

// ── Setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();

  // Default: feature enabled, no announcement available
  mockGetConfig.mockReturnValue({ announcements: true } as ReturnType<
    typeof getConfig
  >);
  mockUseLatestAnnouncement.mockReturnValue({
    announcement: null,
    isLoading: false,
  });
});

afterEach(cleanup);

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AnnouncementModalTrigger", () => {
  describe("when announcements feature flag is disabled", () => {
    it("renders nothing without calling any hooks", () => {
      mockGetConfig.mockReturnValue({ announcements: false } as ReturnType<
        typeof getConfig
      >);

      render(<AnnouncementModalTrigger />);

      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
      // The inner component (which calls the hook) must not have mounted
      expect(mockUseLatestAnnouncement).not.toHaveBeenCalled();
    });
  });

  describe("when no announcement is available", () => {
    it("renders nothing", () => {
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: null,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when an announcement is available and already seen", () => {
    it("renders nothing when the stored hash matches", () => {
      const ann = makeAnnouncement();
      localStorage.setItem("announcement", computeHash(ann));
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when a new (unseen) announcement is available", () => {
    it("shows the modal", () => {
      const ann = makeAnnouncement();
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      expect(screen.getByTestId("announcement-modal")).toBeInTheDocument();
    });

    it("passes the announcement title to the modal", () => {
      const ann = makeAnnouncement({ title: "Important Update" });
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      expect(screen.getByTestId("modal-title")).toHaveTextContent(
        "Important Update",
      );
    });

    it("shows the modal when localStorage has a hash for a different announcement", () => {
      const old = makeAnnouncement({ uuid: "old-uuid", title: "Old" });
      localStorage.setItem("announcement", computeHash(old));

      const fresh = makeAnnouncement({ uuid: "new-uuid", title: "New" });
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: fresh,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      expect(screen.getByTestId("announcement-modal")).toBeInTheDocument();
    });
  });

  describe("on modal close", () => {
    it("hides the modal after it is closed", async () => {
      const user = userEvent.setup();
      const ann = makeAnnouncement();
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);
      expect(screen.getByTestId("announcement-modal")).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(
          screen.queryByTestId("announcement-modal"),
        ).not.toBeInTheDocument();
      });
    });

    it("stores the announcement hash in localStorage when closed", async () => {
      const user = userEvent.setup();
      const ann = makeAnnouncement();
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(localStorage.getItem("announcement")).toBe(computeHash(ann));
      });
    });

    it("does not re-show the modal after dismiss within the same render", async () => {
      const user = userEvent.setup();
      const ann = makeAnnouncement();
      mockUseLatestAnnouncement.mockReturnValue({
        announcement: ann,
        isLoading: false,
      });

      render(<AnnouncementModalTrigger />);

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(
          screen.queryByTestId("announcement-modal"),
        ).not.toBeInTheDocument();
      });

      // Hash is now stored — a fresh render should not show the modal
      cleanup();
      render(<AnnouncementModalTrigger />);
      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });
});
