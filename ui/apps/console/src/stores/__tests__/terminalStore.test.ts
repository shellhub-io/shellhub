import { describe, it, expect, beforeEach } from "vitest";
import { useTerminalStore } from "../terminalStore";

function openTerminal() {
  useTerminalStore.getState().open({
    deviceUid: "dev-1",
    deviceName: "dev-1",
    username: "root",
    password: "",
  });
  return useTerminalStore.getState().sessions.at(-1)!.id;
}

const recording = { id: "session-1", title: "dev-1", logs: "cast" };

const shownRecordings = () =>
  useTerminalStore
    .getState()
    .recordings.filter((r) => r.shown)
    .map((r) => r.id);

beforeEach(() => {
  useTerminalStore.setState({
    sessions: [],
    recordings: [],
    restoreAfterNavigation: null,
  });
});

describe("terminalStore recordings", () => {
  it("puts the terminals away when a recording opens", () => {
    openTerminal();

    useTerminalStore.getState().openRecording(recording);

    expect(useTerminalStore.getState().sessions[0].state).toBe("minimized");
    expect(shownRecordings()).toEqual(["session-1"]);
  });

  it("brings an open recording forward instead of opening it twice", () => {
    const { openRecording, minimizeAll } = useTerminalStore.getState();
    openRecording(recording);
    minimizeAll();

    openRecording(recording);

    expect(useTerminalStore.getState().recordings).toHaveLength(1);
    expect(shownRecordings()).toEqual(["session-1"]);
  });

  it("puts the recording away when a terminal comes forward", () => {
    const id = openTerminal();
    useTerminalStore.getState().openRecording(recording);

    useTerminalStore.getState().restore(id);

    expect(shownRecordings()).toEqual([]);
  });

  it("puts the recording away with everything else", () => {
    useTerminalStore.getState().openRecording(recording);

    useTerminalStore.getState().minimizeAll();

    expect(shownRecordings()).toEqual([]);
  });
});
