import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Announcement } from "@/client";

// ── Dependency mocks ──────────────────────────────────────────────────────────

// Silence the CSS import — jsdom cannot process it
vi.mock("../AnnouncementModal.css", () => ({}));

// Mock the focus trap so jsdom focus state doesn't interfere
vi.mock("@/hooks/useFocusTrap", () => ({ useFocusTrap: vi.fn() }));

// jsdom doesn't implement showModal/close — stub them
HTMLDialogElement.prototype.showModal = vi.fn(function (
  this: HTMLDialogElement,
) {
  this.setAttribute("open", "");
});
HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
  this.removeAttribute("open");
});

// Tiptap uses real DOM APIs that aren't fully available in jsdom.
// Mock the whole editor so AnnouncementContent renders a stable placeholder.
vi.mock("@tiptap/react", () => ({
  useEditor: vi.fn(() => null),
  EditorContent: () => null,
}));

vi.mock("@tiptap/starter-kit", () => ({ default: {} }));
vi.mock("@tiptap/extension-link", () => ({
  default: { configure: vi.fn(() => ({})) },
}));
vi.mock("@tiptap/extension-image", () => ({
  default: { configure: vi.fn(() => ({})) },
}));
vi.mock("@tiptap/markdown", () => ({ Markdown: {} }));
vi.mock("@/utils/url", () => ({ isAllowedUrl: vi.fn(() => true) }));

import AnnouncementModal from "../AnnouncementModal";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeAnnouncement(overrides: Partial<Announcement> = {}): Announcement {
  return {
    uuid: "ann-uuid-1",
    title: "New Feature Released",
    content: "## Hello\nThis is the content.",
    date: "2024-06-15T00:00:00Z",
    ...overrides,
  };
}

function renderModal({
  open = true,
  onClose = vi.fn(),
  announcement = makeAnnouncement(),
}: {
  open?: boolean;
  onClose?: () => void;
  announcement?: Announcement;
} = {}) {
  return {
    onClose,
    ...render(
      <AnnouncementModal
        open={open}
        onClose={onClose}
        announcement={announcement}
      />,
    ),
  };
}

// ── Setup / teardown ──────────────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(cleanup);

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AnnouncementModal", () => {
  describe("when open=false", () => {
    it("renders nothing", () => {
      renderModal({ open: false });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("when open=true", () => {
    it("renders the dialog element", () => {
      renderModal();
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("shows the announcement title", () => {
      renderModal({
        announcement: makeAnnouncement({ title: "Big Announcement" }),
      });
      expect(screen.getByText("Big Announcement")).toBeInTheDocument();
    });

    it("shows the formatted announcement date", () => {
      renderModal({
        announcement: makeAnnouncement({ date: "2024-06-15T12:00:00Z" }),
      });
      // formatDateShort renders "Jun 15, 2024"
      expect(screen.getByText("Jun 15, 2024")).toBeInTheDocument();
    });

    it("renders the close button with correct aria-label", () => {
      renderModal();
      expect(
        screen.getByRole("button", { name: /close announcement/i }),
      ).toBeInTheDocument();
    });

    it("renders the 'Got it' button", () => {
      renderModal();
      expect(
        screen.getByRole("button", { name: /got it/i }),
      ).toBeInTheDocument();
    });
  });

  describe("accessibility", () => {
    it("dialog is labelled by the title element", () => {
      renderModal({
        announcement: makeAnnouncement({ title: "My Announcement" }),
      });
      const dialog = screen.getByRole("dialog");
      const labelledById = dialog.getAttribute("aria-labelledby");
      expect(labelledById).toBeTruthy();

      const titleEl = document.getElementById(labelledById!);
      expect(titleEl).not.toBeNull();
      expect(titleEl!.textContent).toBe("My Announcement");
    });
  });

  describe("close interactions", () => {
    it("calls onClose when the close button is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderModal();

      await user.click(
        screen.getByRole("button", { name: /close announcement/i }),
      );

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when 'Got it' is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderModal();

      await user.click(screen.getByRole("button", { name: /got it/i }));

      expect(onClose).toHaveBeenCalledOnce();
    });
  });
});
