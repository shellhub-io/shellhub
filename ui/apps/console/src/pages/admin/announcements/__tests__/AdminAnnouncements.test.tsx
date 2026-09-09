import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import AdminAnnouncements from "../index";
import type { AnnouncementShort } from "@/client/model";
import { createTestWrapper } from "@/tests/wrapper";
import { mockAnnouncement } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

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

let lastRequestUrl: URL | null;

function setAnnouncements(
  items: ReturnType<typeof mockAnnouncement>[],
  total?: number,
) {
  server.use(
    http.get("*/admin/api/announcements", ({ request }) => {
      lastRequestUrl = new URL(request.url);
      return jsonWithTotal(items, total ?? items.length);
    }),
  );
}

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
    lastRequestUrl = null;
    useAuthStore.setState({ isAdmin: true });
    setAnnouncements([]);
  });

  describe("loading state", () => {
    it("renders the loading spinner while loading", () => {
      server.use(
        http.get("*/admin/api/announcements", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
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
  });

  describe("announcement rows", () => {
    it("renders a row for each returned announcement", async () => {
      setAnnouncements([
        mockAnnouncement({
          uuid: "uuid-a1b2",
          title: "Alpha Announcement",
        }),
        mockAnnouncement({
          uuid: "uuid-c3d4",
          title: "Beta Announcement",
        }),
      ]);
      renderPage();
      expect(await screen.findByText("Alpha Announcement")).toBeInTheDocument();
      expect(screen.getByText("Beta Announcement")).toBeInTheDocument();
    });

    it("renders a truncated UUID chip for each row", async () => {
      setAnnouncements(
        [
          mockAnnouncement({
            uuid: "abcdef12-0000-0000-0000-000000000000",
          }),
        ],
        1,
      );
      renderPage();
      expect(await screen.findByText("abcdef12")).toBeInTheDocument();
    });
  });

  describe("navigation", () => {
    it("navigates to the announcement detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      setAnnouncements(
        [
          mockAnnouncement({
            uuid: "uuid-nav1",
            title: "Clickable Announcement",
          }),
        ],
        1,
      );
      renderPage();

      await user.click(await screen.findByText("Clickable Announcement"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/announcements/uuid-nav1",
      );
    });

    it("navigates to the edit page when the edit button is clicked", async () => {
      const user = userEvent.setup();
      setAnnouncements(
        [
          mockAnnouncement({
            uuid: "uuid-edit1",
            title: "Editable Announcement",
          }),
        ],
        1,
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
      setAnnouncements([
        mockAnnouncement({ uuid: "uuid-edit2", title: "Edit Only" }),
      ]);
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
      setAnnouncements([mockAnnouncement({ title: "Target Announcement" })]);
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

    it("does not navigate when delete button is clicked (stopPropagation)", async () => {
      const user = userEvent.setup();
      setAnnouncements([mockAnnouncement({ title: "No Nav Announcement" })]);
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
    it("renders an error alert when the API returns an error", async () => {
      server.use(
        http.get("*/admin/api/announcements", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      expect(await screen.findByRole("alert")).toHaveTextContent(
        "Something went wrong on our side. Try again.",
      );
    });
  });

  describe("pagination", () => {
    it("does not render pagination when there is only one page", async () => {
      setAnnouncements([mockAnnouncement()]);
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
      setAnnouncements(manyAnnouncements, 25);
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
      setAnnouncements(manyAnnouncements, 25);
      renderPage();
      expect(await screen.findByText("25 announcements")).toBeInTheDocument();
    });
  });

  describe("delete-last-row page decrement (usePaginatedListState)", () => {
    it("decrements page from 2 to 1 via URL when deleting the last item on a page", async () => {
      const user = userEvent.setup();

      setAnnouncements([mockAnnouncement({ title: "Last Item" })], 11);

      renderPage(["/?page=2"]);

      await waitFor(() => {
        expect(lastRequestUrl).not.toBeNull();
        expect(lastRequestUrl!.searchParams.get("page")).toBe("2");
      });

      await user.click(
        await screen.findByRole("button", { name: "Delete Last Item" }),
      );
      await waitFor(() => screen.getByRole("dialog"));

      lastRequestUrl = null;

      await user.click(screen.getByRole("button", { name: "Confirm delete" }));

      await waitFor(() => {
        expect(lastRequestUrl).not.toBeNull();
        expect(lastRequestUrl!.searchParams.get("page")).toBe("1");
      });
    });
  });
});
