import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Announcement } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockAnnouncement, mockAnnouncementFull } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import AnnouncementModalTrigger from "../AnnouncementModalTrigger";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listAnnouncements: vi.fn(),
    getAnnouncement: vi.fn(),
  }),
);

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
        <button type="button" onClick={onClose}>
          Close
        </button>
      </div>
    ) : null,
}));

const mockGetConfig = vi.mocked(getConfig);

function computeHash(announcement: Announcement): string {
  return btoa(JSON.stringify(announcement));
}

function setupAnnouncement(overrides: Partial<Announcement> = {}) {
  const short = mockAnnouncement(overrides);
  const full = mockAnnouncementFull({
    ...overrides,
    uuid: short.uuid,
    title: short.title,
    date: short.date,
  });
  sdk.listAnnouncements.mockResolvedValue(paginatedResponse([short]));
  sdk.getAnnouncement.mockResolvedValue(mockSdkResponse(full));
  return full;
}

function setupNoAnnouncement() {
  sdk.listAnnouncements.mockResolvedValue(paginatedResponse([]));
}

function renderTrigger() {
  return render(<AnnouncementModalTrigger />, {
    wrapper: createTestWrapper(),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockGetConfig.mockReturnValue({ ...defaultConfig, announcements: true });
  setupNoAnnouncement();
});

describe("AnnouncementModalTrigger", () => {
  describe("when announcements feature flag is disabled", () => {
    it("renders nothing without calling any hooks", () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig, announcements: false });

      renderTrigger();

      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
      expect(sdk.listAnnouncements).not.toHaveBeenCalled();
    });
  });

  describe("when no announcement is available", () => {
    it("renders nothing", () => {
      renderTrigger();

      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when an announcement is available and already seen", () => {
    it("renders nothing when the stored hash matches", async () => {
      const ann = setupAnnouncement();
      localStorage.setItem("announcement", computeHash(ann));

      renderTrigger();

      await waitFor(() => expect(sdk.getAnnouncement).toHaveBeenCalled());
      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });

  describe("when a new (unseen) announcement is available", () => {
    it("shows the modal", async () => {
      setupAnnouncement();

      renderTrigger();

      expect(
        await screen.findByTestId("announcement-modal"),
      ).toBeInTheDocument();
    });

    it("passes the announcement title to the modal", async () => {
      setupAnnouncement({ title: "Important Update" });

      renderTrigger();

      expect(await screen.findByTestId("modal-title")).toHaveTextContent(
        "Important Update",
      );
    });

    it("shows the modal when localStorage has a hash for a different announcement", async () => {
      const old = mockAnnouncementFull({ uuid: "old-uuid", title: "Old" });
      localStorage.setItem("announcement", computeHash(old));

      setupAnnouncement({ uuid: "new-uuid", title: "New" });

      renderTrigger();

      expect(
        await screen.findByTestId("announcement-modal"),
      ).toBeInTheDocument();
    });
  });

  describe("on modal close", () => {
    it("hides the modal after it is closed", async () => {
      const user = userEvent.setup();
      setupAnnouncement();

      renderTrigger();
      expect(
        await screen.findByTestId("announcement-modal"),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(
          screen.queryByTestId("announcement-modal"),
        ).not.toBeInTheDocument();
      });
    });

    it("stores the announcement hash in localStorage when closed", async () => {
      const user = userEvent.setup();
      const ann = setupAnnouncement();

      renderTrigger();
      expect(
        await screen.findByTestId("announcement-modal"),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(localStorage.getItem("announcement")).toBe(computeHash(ann));
      });
    });

    it("does not re-show the modal after dismiss within the same render", async () => {
      const user = userEvent.setup();
      setupAnnouncement();

      renderTrigger();
      expect(
        await screen.findByTestId("announcement-modal"),
      ).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: /close/i }));

      await waitFor(() => {
        expect(
          screen.queryByTestId("announcement-modal"),
        ).not.toBeInTheDocument();
      });

      cleanup();
      renderTrigger();

      await waitFor(() => expect(sdk.listAnnouncements).toHaveBeenCalled());
      expect(
        screen.queryByTestId("announcement-modal"),
      ).not.toBeInTheDocument();
    });
  });
});
