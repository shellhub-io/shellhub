import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import AdminAnnouncements from "../index";
import type { AnnouncementShort } from "@/client";
import { makeSdkError, paginatedResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockAnnouncement } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listAnnouncementsAdmin: vi.fn(),
  }),
);

vi.mock("../DeleteAnnouncementDialog", () => ({
  default: ({
    open,
    onClose,
    onDeleted,
    announcement,
  }: {
    open: boolean;
    onClose: () => void;
    onDeleted?: () => void;
    announcement: AnnouncementShort | null;
  }) => {
    if (!open || !announcement) return null;
    return (
      <div role="dialog" aria-label={`Delete ${announcement.title}`}>
        <button type="button" onClick={onClose}>
          Cancel delete
        </button>
        <button
          type="button"
          onClick={() => {
            onClose();
            onDeleted?.();
          }}
        >
          Confirm delete
        </button>
      </div>
    );
  },
}));

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminAnnouncements />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminAnnouncements", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true });
    sdk.listAnnouncementsAdmin.mockResolvedValue(paginatedResponse([]));
  });

  describe("rendering", () => {
    it('renders the page heading "Announcements"', async () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Announcements" }),
      ).toBeInTheDocument();
    });

    it("renders the Announcements table", async () => {
      renderPage();
      await screen.findByText("No announcements found");
      expect(
        screen.getByRole("table", { name: "Announcements" }),
      ).toBeInTheDocument();
    });

    it("renders the column headers", async () => {
      renderPage();
      await screen.findByText("No announcements found");
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
      sdk.listAnnouncementsAdmin.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
    });

    it("renders 'Loading announcements...' text while loading", () => {
      sdk.listAnnouncementsAdmin.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByText("Loading announcements...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it("renders 'No announcements found' when the list is empty", async () => {
      renderPage();
      expect(
        await screen.findByText("No announcements found"),
      ).toBeInTheDocument();
    });

    it("does not render announcement rows when the list is empty", async () => {
      renderPage();
      await screen.findByText("No announcements found");
      expect(
        screen.queryByRole("button", { name: /edit/i }),
      ).not.toBeInTheDocument();
    });
  });

  describe("announcement rows", () => {
    it("renders a row for each returned announcement", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([
          mockAnnouncement({
            uuid: "uuid-a1b2",
            title: "Alpha Announcement",
          }),
          mockAnnouncement({
            uuid: "uuid-c3d4",
            title: "Beta Announcement",
          }),
        ]),
      );
      renderPage();
      expect(await screen.findByText("Alpha Announcement")).toBeInTheDocument();
      expect(screen.getByText("Beta Announcement")).toBeInTheDocument();
    });

    it("renders a truncated UUID chip for each row", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(
          [
            mockAnnouncement({
              uuid: "abcdef12-0000-0000-0000-000000000000",
            }),
          ],
          1,
        ),
      );
      renderPage();
      expect(await screen.findByText("abcdef12")).toBeInTheDocument();
    });

    it("renders a formatted date for each row", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([
          mockAnnouncement({ date: "2024-06-01T10:00:00.000Z" }),
        ]),
      );
      renderPage();
      const dateCell = await screen.findByText(/\d{4}/);
      expect(dateCell).toBeInTheDocument();
    });

    it("renders an edit button for each row", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "My Announcement" })]),
      );
      renderPage();
      expect(
        await screen.findByRole("button", { name: "Edit My Announcement" }),
      ).toBeInTheDocument();
    });

    it("renders a delete button for each row", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "My Announcement" })]),
      );
      renderPage();
      expect(
        await screen.findByRole("button", { name: "Delete My Announcement" }),
      ).toBeInTheDocument();
    });
  });

  describe("navigation", () => {
    it("navigates to the announcement detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(
          [
            mockAnnouncement({
              uuid: "uuid-nav1",
              title: "Clickable Announcement",
            }),
          ],
          1,
        ),
      );
      renderPage();

      await user.click(await screen.findByText("Clickable Announcement"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/announcements/uuid-nav1",
      );
    });

    it("navigates to the edit page when the edit button is clicked", async () => {
      const user = userEvent.setup();
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(
          [
            mockAnnouncement({
              uuid: "uuid-edit1",
              title: "Editable Announcement",
            }),
          ],
          1,
        ),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", {
          name: "Edit Editable Announcement",
        }),
      );
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/announcements/uuid-edit1/edit",
      );
    });

    it("does not navigate to the detail page when edit button is clicked", async () => {
      const user = userEvent.setup();
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([
          mockAnnouncement({ uuid: "uuid-edit2", title: "Edit Only" }),
        ]),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", { name: "Edit Edit Only" }),
      );

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
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "Target Announcement" })]),
      );
      renderPage();

      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

      await user.click(
        await screen.findByRole("button", {
          name: "Delete Target Announcement",
        }),
      );

      await waitFor(() =>
        expect(screen.getByRole("dialog")).toBeInTheDocument(),
      );
    });

    it("closes the DeleteAnnouncementDialog when cancel is clicked inside it", async () => {
      const user = userEvent.setup();
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "Target Announcement" })]),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", {
          name: "Delete Target Announcement",
        }),
      );
      await waitFor(() => screen.getByRole("dialog"));

      await user.click(screen.getByRole("button", { name: "Cancel delete" }));

      await waitFor(() =>
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
      );
    });

    it("does not navigate when delete button is clicked (stopPropagation)", async () => {
      const user = userEvent.setup();
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "No Nav Announcement" })]),
      );
      renderPage();

      await user.click(
        await screen.findByRole("button", {
          name: "Delete No Nav Announcement",
        }),
      );

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      sdk.listAnnouncementsAdmin.mockRejectedValue(makeSdkError(500));
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
    });

    it("renders the error message text", async () => {
      sdk.listAnnouncementsAdmin.mockRejectedValue(makeSdkError(500));
      renderPage();
      expect(
        await screen.findByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("pagination", () => {
    it("does not render pagination when there is only one page", async () => {
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement()]),
      );
      renderPage();
      await screen.findByText("Welcome to ShellHub");
      expect(
        screen.queryByRole("button", { name: "Previous page" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Next page" }),
      ).not.toBeInTheDocument();
    });

    it("renders pagination controls when there are multiple pages", async () => {
      const manyAnnouncements = Array.from({ length: 10 }, (_, i) =>
        mockAnnouncement({ uuid: `uuid-${i}`, title: `Ann ${i}` }),
      );
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(manyAnnouncements, 25),
      );
      renderPage();
      expect(
        await screen.findByRole("button", { name: "Previous page" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Next page" }),
      ).toBeInTheDocument();
    });

    it("renders the item count label in the pagination area", async () => {
      const manyAnnouncements = Array.from({ length: 10 }, (_, i) =>
        mockAnnouncement({ uuid: `uuid-${i}`, title: `Ann ${i}` }),
      );
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(manyAnnouncements, 25),
      );
      renderPage();
      expect(await screen.findByText("25 announcements")).toBeInTheDocument();
    });
  });

  describe("URL hydration (usePaginatedListState)", () => {
    it("passes page=2 to the SDK when URL has ?page=2", async () => {
      renderPage(["/?page=2"]);
      await waitFor(() => {
        expect(sdk.listAnnouncementsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2 }),
          }),
        );
      });
    });

    it("passes page=1 to the SDK when URL has no page param", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(sdk.listAnnouncementsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
    });
  });

  describe("URL writes (usePaginatedListState)", () => {
    it("passes page=2 to the SDK when the user clicks Next page", async () => {
      const user = userEvent.setup();
      const manyAnnouncements = Array.from({ length: 10 }, (_, i) =>
        mockAnnouncement({ uuid: `uuid-${i}`, title: `Ann ${i}` }),
      );
      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse(manyAnnouncements, 30),
      );
      renderPage();

      await screen.findByText("Ann 0");

      await user.click(screen.getByRole("button", { name: "Next page" }));

      await waitFor(() => {
        expect(sdk.listAnnouncementsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2 }),
          }),
        );
      });
    });
  });

  describe("delete-last-row page decrement (usePaginatedListState)", () => {
    it("decrements page from 2 to 1 via URL when deleting the last item on a page", async () => {
      const user = userEvent.setup();

      sdk.listAnnouncementsAdmin.mockResolvedValue(
        paginatedResponse([mockAnnouncement({ title: "Last Item" })], 11),
      );

      renderPage(["/?page=2"]);

      await waitFor(() => {
        expect(sdk.listAnnouncementsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2 }),
          }),
        );
      });

      await user.click(
        await screen.findByRole("button", { name: "Delete Last Item" }),
      );
      await waitFor(() => screen.getByRole("dialog"));

      sdk.listAnnouncementsAdmin.mockClear();

      await user.click(screen.getByRole("button", { name: "Confirm delete" }));

      await waitFor(() => {
        expect(sdk.listAnnouncementsAdmin).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
    });
  });
});
