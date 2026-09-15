import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Announcement } from "@/client/model";

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

beforeEach(() => {
  vi.clearAllMocks();
});

describe("AnnouncementModal", () => {
  it("renders nothing when open=false", () => {
    renderModal({ open: false });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows the formatted announcement date", () => {
    renderModal({
      announcement: makeAnnouncement({ date: "2024-06-15T12:00:00Z" }),
    });
    expect(screen.getByText("Jun 15, 2024")).toBeInTheDocument();
  });

  it("labels the dialog with the announcement title", () => {
    renderModal({
      announcement: makeAnnouncement({ title: "My Announcement" }),
    });
    const labelledById = screen
      .getByRole("dialog")
      .getAttribute("aria-labelledby");
    expect(labelledById).toBeTruthy();

    const titleEl = document.getElementById(labelledById!);
    expect(titleEl).not.toBeNull();
    expect(titleEl!.textContent).toBe("My Announcement");
  });

  it.each([/close announcement/i, /got it/i])(
    "calls onClose when the %s button is clicked",
    async (name) => {
      const user = userEvent.setup();
      const { onClose } = renderModal();

      await user.click(screen.getByRole("button", { name }));

      expect(onClose).toHaveBeenCalledOnce();
    },
  );
});
