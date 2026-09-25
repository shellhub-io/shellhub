import { describe, it, expect, beforeEach } from "vitest";
import { orderedWindows, useTerminalStore } from "../terminalStore";

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
    windowOrder: [],
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

describe("terminalStore window order", () => {
  const order = () => orderedWindows(useTerminalStore.getState());

  it("lines terminals and recordings up in the order they opened", () => {
    const terminal = openTerminal();
    useTerminalStore.getState().openRecording(recording);

    expect(order()).toEqual([terminal, "session-1"]);
  });

  it("moves a recording in front of a terminal", () => {
    const terminal = openTerminal();
    useTerminalStore.getState().openRecording(recording);

    useTerminalStore.getState().moveWindow("session-1", 0);

    expect(order()).toEqual(["session-1", terminal]);
  });

  it("drops a closed window from the order, so reopening it goes to the end", () => {
    const { openRecording, closeRecording } = useTerminalStore.getState();
    openRecording(recording);
    const terminal = openTerminal();

    closeRecording("session-1");
    openRecording(recording);

    expect(order()).toEqual([terminal, "session-1"]);
  });

  it("drops a closed terminal from the stored order", () => {
    const first = openTerminal();
    const second = openTerminal();

    useTerminalStore.getState().close(first);

    expect(useTerminalStore.getState().windowOrder).toEqual([second]);
  });

  it("drops a terminal closed to reconnect from the stored order", () => {
    const first = openTerminal();
    const second = openTerminal();

    useTerminalStore.getState().closeAndReconnect(first);

    expect(useTerminalStore.getState().windowOrder).toEqual([second]);
  });

  it("puts a new window after every open one, even those missing from the stored order", () => {
    const earlier = openTerminal();
    useTerminalStore.setState({ windowOrder: [] });

    const later = openTerminal();

    expect(order()).toEqual([earlier, later]);
  });
});
