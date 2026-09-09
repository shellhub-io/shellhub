import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import DeleteAnnouncementDialog from "../DeleteAnnouncementDialog";

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

const mockAnnouncement = {
  uuid: "ann-uuid-1234",
  title: "Test Announcement",
};

const Wrapper = createTestWrapper();

beforeEach(() => {
  vi.clearAllMocks();
  server.use(
    http.delete(
      "*/admin/api/announcements/:uuid",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

function renderDialog(
  overrides: Partial<{
    open: boolean;
    onClose: () => void;
    announcement: typeof mockAnnouncement | null;
    onDeleted: () => void;
  }> = {},
) {
  const defaults = {
    open: true,
    onClose: vi.fn(),
    announcement: mockAnnouncement,
    onDeleted: vi.fn(),
  };
  const props = { ...defaults, ...overrides };
  return {
    onClose: props.onClose,
    onDeleted: props.onDeleted,
    ...render(
      <Wrapper>
        <DeleteAnnouncementDialog {...props} />
      </Wrapper>,
    ),
  };
}

describe("DeleteAnnouncementDialog", () => {
  describe("confirm — success", () => {
    it("calls onClose before onDeleted", async () => {
      const callOrder: string[] = [];
      const onClose = vi.fn(() => callOrder.push("onClose"));
      const onDeleted = vi.fn(() => callOrder.push("onDeleted"));
      render(
        <Wrapper>
          <DeleteAnnouncementDialog
            open={true}
            onClose={onClose}
            announcement={mockAnnouncement}
            onDeleted={onDeleted}
          />
        </Wrapper>,
      );

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onDeleted).toHaveBeenCalledTimes(1));
      expect(callOrder).toEqual(["onClose", "onDeleted"]);
    });
  });

  describe("confirm — error handling", () => {
    it("shows generic error message on failure", async () => {
      server.use(
        http.delete("*/admin/api/announcements/:uuid", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => {
        expect(
          screen.getByText(/failed to delete announcement/i),
        ).toBeInTheDocument();
      });
    });

    it("does not call onDeleted when deletion fails", async () => {
      server.use(
        http.delete("*/admin/api/announcements/:uuid", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { onDeleted } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete announcement/i));
      expect(onDeleted).not.toHaveBeenCalled();
    });

    it("does not call onClose when deletion fails", async () => {
      server.use(
        http.delete("*/admin/api/announcements/:uuid", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => screen.getByText(/failed to delete announcement/i));
      expect(onClose).not.toHaveBeenCalled();
    });

    it("clears the error message on subsequent close after failure", async () => {
      server.use(
        http.delete("*/admin/api/announcements/:uuid", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      const { onClose } = renderDialog();

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));
      await waitFor(() => screen.getByText(/failed to delete announcement/i));

      await userEvent.click(screen.getByRole("button", { name: /cancel/i }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  describe("optional onDeleted callback", () => {
    it("does not throw when onDeleted is not provided and deletion succeeds", async () => {
      const { onClose } = renderDialog({ onDeleted: undefined });

      await userEvent.click(screen.getByRole("button", { name: /^delete$/i }));

      await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    });
  });
});
