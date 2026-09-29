import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockRecordingMeta } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useRecordingsStore } from "@/stores/recordingsStore";
import RecordingActionsMenu from "../RecordingActionsMenu";

async function openMenu({
  local,
  recorded,
  onDownload = () => {},
  loaded = false,
}: {
  local: boolean;
  recorded: boolean;
  onDownload?: () => void;
  loaded?: boolean;
}) {
  useRecordingsStore.setState({
    recordings: local ? [mockRecordingMeta()] : [],
  });
  const user = userEvent.setup();
  render(
    <RecordingActionsMenu
      sessionUid="session-1"
      recorded={recorded}
      onDownload={onDownload}
      loaded={loaded}
    />,
    { wrapper: createTestWrapper() },
  );
  await user.click(screen.getByRole("button", { name: "Recording actions" }));
  return user;
}

function serveDelete(status: number) {
  server.use(
    http.delete(
      "*/api/sessions/:uid/records/:seat",
      () => new HttpResponse(null, { status }),
    ),
  );
}

async function confirmDelete(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("menuitem", { name: "Delete recording" }));
  await user.click(
    within(screen.getByRole("dialog")).getByRole("button", {
      name: "Delete recording",
    }),
  );
}

const restricted = (name: string) =>
  screen.getByRole("menuitem", { name }).closest("[aria-disabled]") !== null;

beforeEach(() => {
  seedAuthStore({ role: "operator" });
});

describe("RecordingActionsMenu", () => {
  it("lets a member without session permissions act on the copy their browser holds", async () => {
    await openMenu({ local: true, recorded: false });

    expect(restricted("Download recording")).toBe(false);
    expect(restricted("Delete recording")).toBe(false);
  });

  it("holds back a member without session permissions from the server's copy", async () => {
    await openMenu({ local: false, recorded: true });

    expect(restricted("Download recording")).toBe(true);
    expect(restricted("Delete recording")).toBe(true);
  });

  it("downloads the browser's copy but guards deleting the server's when both exist", async () => {
    await openMenu({ local: true, recorded: true });

    expect(restricted("Download recording")).toBe(false);
    expect(restricted("Delete recording")).toBe(true);
  });

  it("lets them download a recording the caller already holds, even when only the server has a copy", async () => {
    await openMenu({ local: false, recorded: true, loaded: true });

    expect(restricted("Download recording")).toBe(false);
    expect(restricted("Delete recording")).toBe(true);
  });

  describe("as an administrator", () => {
    beforeEach(() => {
      seedAuthStore({ role: "administrator" });
    });

    it("downloads when Download recording is picked", async () => {
      const onDownload = vi.fn();
      const user = await openMenu({ local: false, recorded: true, onDownload });

      await user.click(
        screen.getByRole("menuitem", { name: "Download recording" }),
      );

      expect(onDownload).toHaveBeenCalledOnce();
    });

    it("closes the confirmation once the recording is deleted", async () => {
      serveDelete(200);
      const user = await openMenu({ local: false, recorded: true });

      await confirmDelete(user);

      await waitFor(() =>
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
      );
    });

    it("keeps the confirmation open with the error when the delete fails", async () => {
      serveDelete(500);
      const user = await openMenu({ local: false, recorded: true });

      await confirmDelete(user);

      expect(
        await within(screen.getByRole("dialog")).findByText(
          "Failed to delete recording.",
        ),
      ).toBeInTheDocument();
    });
  });
});
