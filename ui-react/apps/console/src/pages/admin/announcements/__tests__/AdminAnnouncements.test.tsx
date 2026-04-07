import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { AnnouncementShort } from "../../../../client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("../../../../hooks/useAdminAnnouncements", () => ({
  useAdminAnnouncements: vi.fn(),
}));

// useNavigate is used inside AnnouncementRow — mock at the module level.
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// DeleteAnnouncementDialog pulls in the mutation hook and ConfirmDialog.
// Mock the whole component so list-page tests stay focused on rendering/nav.
vi.mock("../DeleteAnnouncementDialog", () => ({
  default: ({
    open,
    onClose,
    announcement,
  }: {
    open: boolean;
    onClose: () => void;
    announcement: AnnouncementShort | null;
  }) => {
    if (!open || !announcement) return null;
    return (
      <div role="dialog" aria-label={`Delete ${announcement.title}`}>
        <button onClick={onClose}>Cancel delete</button>
      </div>
    );
  },
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useAdminAnnouncements } from "../../../../hooks/useAdminAnnouncements";
import AdminAnnouncements from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  announcements: [] as AnnouncementShort[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeAnnouncement(
  overrides: Partial<AnnouncementShort> = {},
): AnnouncementShort {
  return {
    uuid: "uuid-0001",
    title: "Welcome to ShellHub",
    date: "2024-06-01T10:00:00.000Z",
    ...overrides,
  };
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminAnnouncements />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminAnnouncements", () => {
  beforeEach(() => {
    vi.mocked(useAdminAnnouncements).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
  });

  describe("rendering", () => {
    it('renders the page heading "Announcements"', () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Announcements" }),
      ).toBeInTheDocument();
    });

    it("renders the Announcements table", () => {
      renderPage();
      expect(
        screen.getByRole("table", { name: "Announcements" }),
      ).toBeInTheDocument();
    });

    it("renders the column headers", () => {
      renderPage();
      expect(
        screen.getByRole("columnheader", { name: "UUID" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("columnheader", { name: "Title" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("columnheader", { name: "Date" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("columnheader", { name: "Actions" }),
      ).toBeInTheDocument();
    });

    it("renders the 'New' button", () => {
      renderPage();
      expect(screen.getByRole("button", { name: /new/i })).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it("renders the loading spinner with role='status'", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        announcements: [],
      });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("renders 'Loading announcements...' text while loading", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        announcements: [],
      });
      renderPage();
      expect(screen.getByText("Loading announcements...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it("renders 'No announcements found' when the list is empty", () => {
      renderPage();
      expect(screen.getByText("No announcements found")).toBeInTheDocument();
    });

    it("does not render announcement rows when the list is empty", () => {
      renderPage();
      expect(
        screen.queryByRole("button", { name: /edit/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("announcement rows", () => {
    it("renders a row for each returned announcement", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [
          makeAnnouncement({ uuid: "uuid-a1b2", title: "Alpha Announcement" }),
          makeAnnouncement({ uuid: "uuid-c3d4", title: "Beta Announcement" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("Alpha Announcement")).toBeInTheDocument();
      expect(screen.getByText("Beta Announcement")).toBeInTheDocument();
    });

    it("renders a truncated UUID chip for each row", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [
          makeAnnouncement({ uuid: "abcdef12-0000-0000-0000-000000000000" }),
        ],
        totalCount: 1,
      });
      renderPage();
      // The page renders the first 8 chars of the UUID
      expect(screen.getByText("abcdef12")).toBeInTheDocument();
    });

    it("renders a formatted date for each row", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ date: "2024-06-01T10:00:00.000Z" })],
        totalCount: 1,
      });
      renderPage();
      // The date is formatted via toLocaleDateString — just check it's a non-empty
      // string that appears somewhere in the row (exact format is locale-dependent)
      const dateCell = screen.getByText(/\d{4}/);
      expect(dateCell).toBeInTheDocument();
    });

    it("renders an edit button for each row", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ title: "My Announcement" })],
        totalCount: 1,
      });
      renderPage();
      expect(
        screen.getByRole("button", { name: "Edit My Announcement" }),
      ).toBeInTheDocument();
    });

    it("renders a delete button for each row", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ title: "My Announcement" })],
        totalCount: 1,
      });
      renderPage();
      expect(
        screen.getByRole("button", { name: "Delete My Announcement" }),
      ).toBeInTheDocument();
    });
  });

  describe("navigation", () => {
    it("navigates to the announcement detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [
          makeAnnouncement({
            uuid: "uuid-nav1",
            title: "Clickable Announcement",
          }),
        ],
        totalCount: 1,
      });
      renderPage();

      await user.click(screen.getByText("Clickable Announcement"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/announcements/uuid-nav1",
      );
    });

    it("navigates to the edit page when the edit button is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [
          makeAnnouncement({
            uuid: "uuid-edit1",
            title: "Editable Announcement",
          }),
        ],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Edit Editable Announcement" }),
      );
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/announcements/uuid-edit1/edit",
      );
    });

    it("does not navigate to the detail page when edit button is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [
          makeAnnouncement({ uuid: "uuid-edit2", title: "Edit Only" }),
        ],
        totalCount: 1,
      });
      renderPage();

      await user.click(screen.getByRole("button", { name: "Edit Edit Only" }));

      // Only the /edit path should be called — not the bare detail path
      expect(mockNavigate).not.toHaveBeenCalledWith(
        "/admin/announcements/uuid-edit2",
      );
    });

    it("navigates to /admin/announcements/new when 'New' button is clicked", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.click(screen.getByRole("button", { name: /new/i }));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/announcements/new");
    });
  });

  describe("delete action", () => {
    it("opens the DeleteAnnouncementDialog when delete button is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ title: "Target Announcement" })],
        totalCount: 1,
      });
      renderPage();

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

      await user.click(
        screen.getByRole("button", { name: "Delete Target Announcement" }),
      );

      await waitFor(() =>
        expect(screen.getByRole("dialog")).toBeInTheDocument(),
      );
    });

    it("closes the DeleteAnnouncementDialog when cancel is clicked inside it", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ title: "Target Announcement" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Delete Target Announcement" }),
      );
      await waitFor(() => screen.getByRole("dialog"));

      await user.click(screen.getByRole("button", { name: "Cancel delete" }));

      await waitFor(() =>
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
      );
    });

    it("does not navigate when delete button is clicked (stopPropagation)", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement({ title: "No Nav Announcement" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Delete No Nav Announcement" }),
      );

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe("error state", () => {
    it("renders an error alert when the hook returns an error", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("renders the error message text", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  describe("pagination", () => {
    it("does not render pagination when there is only one page", () => {
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: [makeAnnouncement()],
        totalCount: 1,
      });
      renderPage();
      expect(
        screen.queryByRole("button", { name: /prev/i }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /next/i }),
      ).not.toBeInTheDocument();
    });

    it("renders pagination controls when there are multiple pages", () => {
      // PER_PAGE is 10 — need more than 10 items to get >1 page
      const manyAnnouncements = Array.from({ length: 10 }, (_, i) =>
        makeAnnouncement({ uuid: `uuid-${i}`, title: `Ann ${i}` }),
      );
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: manyAnnouncements,
        totalCount: 25,
      });
      renderPage();
      expect(screen.getByRole("button", { name: "Prev" })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Next" })).toBeInTheDocument();
    });

    it("renders the item count label in the pagination area", () => {
      const manyAnnouncements = Array.from({ length: 10 }, (_, i) =>
        makeAnnouncement({ uuid: `uuid-${i}`, title: `Ann ${i}` }),
      );
      vi.mocked(useAdminAnnouncements).mockReturnValue({
        ...defaultHookState,
        announcements: manyAnnouncements,
        totalCount: 25,
      });
      renderPage();
      expect(screen.getByText("25 announcements")).toBeInTheDocument();
    });
  });
});
