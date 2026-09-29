import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { RecordingMeta } from "@/utils/recordings";

const storage = vi.hoisted(() => ({
  supported: true,
  cleared: vi.fn(() => Promise.resolve()),
}));

vi.mock("@/utils/recordings", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/utils/recordings")>()),
  isRecordingSupported: () => storage.supported,
  listRecordings: () => Promise.resolve([]),
  pruneRecordings: () => Promise.resolve(),
  clearRecordings: storage.cleared,
}));

const { useRecordingsStore } = await import("@/stores/recordingsStore");
const { default: RecordingsPreferences } = await import(
  "../RecordingsPreferences",
);

function recording(id: string, size: number): RecordingMeta {
  return {
    id,
    filename: `${id}.cast`,
    deviceName: "web-01",
    deviceUid: "dev-1",
    username: "root",
    width: 80,
    height: 24,
    durationSec: 10,
    createdAt: 0,
    size,
  };
}

function renderRecordings(recordings: RecordingMeta[] = []) {
  useRecordingsStore.setState({
    recordings,
    refresh: () => Promise.resolve(),
  });
  const user = userEvent.setup();
  render(<RecordingsPreferences />);
  return user;
}

beforeEach(() => {
  localStorage.clear();
  storage.supported = true;
  storage.cleared.mockClear();
  useRecordingsStore.setState({ retentionDays: null });
});

describe("RecordingsPreferences", () => {
  it("keeps recordings for the picked number of days", async () => {
    const user = renderRecordings();

    await user.click(screen.getByRole("radio", { name: "30 days" }));

    expect(useRecordingsStore.getState().retentionDays).toBe(30);
  });

  it("keeps them forever again when forever is picked", async () => {
    useRecordingsStore.setState({ retentionDays: 7 });
    const user = renderRecordings();

    await user.click(screen.getByRole("radio", { name: "Forever" }));

    expect(useRecordingsStore.getState().retentionDays).toBeNull();
  });

  it("asks before deleting every recording, and deletes on confirm", async () => {
    const user = renderRecordings([recording("a", 2048)]);

    await user.click(screen.getByRole("button", { name: "Delete all" }));
    expect(storage.cleared).not.toHaveBeenCalled();

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("The recording kept in this browser");
    await user.click(within(dialog).getByRole("button", { name: "Delete all" }));

    expect(storage.cleared).toHaveBeenCalledOnce();
  });

  it("says so, and offers no retention, where the browser cannot keep recordings", () => {
    storage.supported = false;
    renderRecordings();

    expect(screen.getByText("Not available here")).toBeInTheDocument();
    expect(
      screen.queryByRole("radiogroup", { name: "Keep recordings" }),
    ).not.toBeInTheDocument();
  });
});
