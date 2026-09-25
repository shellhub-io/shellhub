import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Player } from "asciinema-player";
import { useTerminalThemeStore } from "@/stores/terminalThemeStore";
import SessionPlayer from "../SessionPlayer";

vi.mock("asciinema-player", () => ({ create: vi.fn() }));

import { create } from "asciinema-player";

const mockedCreate = vi.mocked(create);

const play = vi.fn<Player["play"]>();
const pause = vi.fn<Player["pause"]>();
const seek = vi.fn<Player["seek"]>();
const dispose = vi.fn<Player["dispose"]>();

let listeners: Record<string, () => void>;
let currentTime: number;
let duration: number | undefined;

const player = {
  el: document.createElement("div"),
  play,
  pause,
  seek,
  dispose,
  getCurrentTime: () => currentTime,
  getDuration: () => duration,
  addEventListener: (event: string, handler: () => void) => {
    listeners[event] = handler;
  },
} as unknown as Player;

function renderPlayer() {
  const user = userEvent.setup();
  render(<SessionPlayer logs="test-logs" />);
  return { user };
}

const emit = (event: string) => act(async () => listeners[event]?.());

const initialTerminalLook = useTerminalThemeStore.getState();

beforeEach(() => {
  useTerminalThemeStore.setState(initialTerminalLook, true);
  listeners = {};
  currentTime = 0;
  duration = 60;
  vi.clearAllMocks();
  seek.mockResolvedValue(undefined);
  mockedCreate.mockReturnValue(player);
});

describe("SessionPlayer", () => {
  describe("mount", () => {
    it("creates a controls-less player from the provided logs and autoplays", () => {
      render(<SessionPlayer logs="my-session-data" />);

      expect(mockedCreate).toHaveBeenCalledWith(
        { data: "my-session-data" },
        expect.any(HTMLElement),
        expect.objectContaining({ controls: false }),
      );
      expect(play).toHaveBeenCalledTimes(1);
    });

    it("disposes the player on unmount", () => {
      const { unmount } = render(<SessionPlayer logs="test-logs" />);
      unmount();

      expect(dispose).toHaveBeenCalledTimes(1);
    });
  });

  describe("playback state", () => {
    beforeEach(() => {
      renderPlayer();
    });

    it("shows the duration and a pause affordance once playback starts", async () => {
      await emit("playing");
      expect(screen.getByText("00:00 / 01:00")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Pause" })).toBeInTheDocument();
    });

    it("advances the time readout as playback progresses", async () => {
      await emit("playing");
      currentTime = 5;
      expect(await screen.findByText("00:05 / 01:00")).toBeInTheDocument();
    });

    it("returns to a play affordance when playback ends", async () => {
      await emit("playing");
      await emit("ended");
      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
    });

    it("shows the time formatted as HH:MM:SS when the duration is one hour or more", async () => {
      duration = 3600;
      await emit("playing");
      expect(screen.getByText("00:00:00 / 01:00:00")).toBeInTheDocument();
    });
  });

  describe("play/pause control", () => {
    it("pauses when the control is clicked while playing", async () => {
      const { user } = renderPlayer();
      await emit("playing");

      await user.click(screen.getByRole("button", { name: "Pause" }));

      expect(pause).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
    });

    it("resumes when the control is clicked while paused", async () => {
      const { user } = renderPlayer();
      play.mockClear();

      await user.click(screen.getByRole("button", { name: "Play" }));

      expect(play).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("button", { name: "Pause" })).toBeInTheDocument();
    });
  });

  describe("keyboard shortcuts", () => {
    it("toggles play/pause with Space", async () => {
      const { user } = renderPlayer();
      await emit("playing");
      pause.mockClear();

      await user.keyboard("[Space]");

      expect(pause).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
    });
  });

  describe("seeking", () => {
    it.each(["{ArrowRight}", "{ArrowLeft}", "5"])(
      "resumes playback after seeking with %s while playing",
      async (keys) => {
        const { user } = renderPlayer();
        await emit("playing");
        play.mockClear();
        pause.mockClear();

        await user.keyboard(keys);

        expect(seek).toHaveBeenCalledTimes(1);
        expect(pause).toHaveBeenCalledTimes(1);
        expect(play).toHaveBeenCalledTimes(1);
      },
    );

    it("seeks without resuming when paused", async () => {
      const { user } = renderPlayer();
      play.mockClear();

      await user.keyboard("{ArrowRight}");

      expect(seek).toHaveBeenCalledTimes(1);
      expect(play).not.toHaveBeenCalled();
    });

    it("steps a frame with comma/period only while paused", async () => {
      const { user } = renderPlayer();

      await user.keyboard(",.");
      expect(seek).toHaveBeenCalledTimes(2);

      seek.mockClear();
      await emit("playing");
      await user.keyboard(",.");
      expect(seek).not.toHaveBeenCalled();
    });
  });

  describe("playback speed", () => {
    it("recreates the player at the current position when speed changes", async () => {
      const { user } = renderPlayer();
      await emit("playing");

      currentTime = 10;
      await screen.findByText("00:10 / 01:00");
      dispose.mockClear();

      await user.click(
        screen.getByRole("button", { name: /Playback speed 1×/ }),
      );

      expect(dispose).toHaveBeenCalledTimes(1);
      expect(mockedCreate).toHaveBeenLastCalledWith(
        { data: "test-logs" },
        expect.any(HTMLElement),
        expect.objectContaining({ speed: 1.5, startAt: 10 }),
      );
    });
  });

  describe("restarting", () => {
    it("stays paused when the speed changes while paused", async () => {
      const { user } = renderPlayer();
      await emit("playing");
      await user.click(screen.getByRole("button", { name: "Pause" }));
      play.mockClear();

      await user.click(
        screen.getByRole("button", { name: /Playback speed 1×/ }),
      );

      expect(play).not.toHaveBeenCalled();
      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
    });

    it("redraws the recording in a new terminal theme", async () => {
      renderPlayer();
      await emit("playing");
      dispose.mockClear();
      mockedCreate.mockClear();

      act(() => {
        useTerminalThemeStore.setState({
          theme: {
            ...initialTerminalLook.theme,
            colors: { background: "#ffffff", foreground: "#000000" },
          },
        });
      });

      await vi.waitFor(() => expect(mockedCreate).toHaveBeenCalledTimes(1));
      expect(dispose).toHaveBeenCalledTimes(1);
    });

    it("recreates once for a burst of font size changes", async () => {
      renderPlayer();
      await emit("playing");
      mockedCreate.mockClear();

      act(() => {
        for (const size of [14, 15, 16, 17]) {
          useTerminalThemeStore.setState({ fontSize: size });
        }
      });

      await vi.waitFor(() =>
        expect(mockedCreate).toHaveBeenLastCalledWith(
          { data: "test-logs" },
          expect.any(HTMLElement),
          expect.objectContaining({ terminalFontSize: "17px" }),
        ),
      );
      expect(mockedCreate).toHaveBeenCalledTimes(1);
    });
  });

  describe("backgrounded before it starts", () => {
    it("stays paused when playback begins behind another tab", async () => {
      const { rerender } = render(<SessionPlayer logs="test-logs" />);
      rerender(<SessionPlayer logs="test-logs" visible={false} />);
      pause.mockClear();

      await emit("playing");

      expect(pause).toHaveBeenCalledTimes(1);
      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
    });
  });

  describe("seeking after the end", () => {
    it("keeps the new position when the player is recreated", async () => {
      const { user } = renderPlayer();
      await emit("playing");
      await emit("ended");
      await user.keyboard("3");
      mockedCreate.mockClear();

      await user.click(
        screen.getByRole("button", { name: /Playback speed 1×/ }),
      );

      expect(mockedCreate).toHaveBeenLastCalledWith(
        { data: "test-logs" },
        expect.any(HTMLElement),
        expect.objectContaining({ startAt: 18 }),
      );
    });
  });

  describe("terminal appearance", () => {
    it("plays in the terminal's font and size", () => {
      useTerminalThemeStore.setState({
        fontFamilyWithFallback: "'JetBrains Mono', monospace",
        fontSize: 17,
      });

      render(<SessionPlayer logs="test-logs" />);

      expect(mockedCreate).toHaveBeenCalledWith(
        { data: "test-logs" },
        expect.any(HTMLElement),
        expect.objectContaining({
          terminalFontFamily: "'JetBrains Mono', monospace",
          terminalFontSize: "17px",
        }),
      );
    });
  });

  describe("controls visibility", () => {
    const bar = () => screen.getByTestId("player-controls");

    it("keeps the controls up while paused", () => {
      renderPlayer();

      expect(bar()).toHaveAttribute("data-state", "shown");
    });

    it("keeps the controls up for a moment once playback starts", async () => {
      renderPlayer();

      await emit("playing");

      expect(bar()).toHaveAttribute("data-state", "shown");
    });

    it("puts the controls away with Escape, even while paused", async () => {
      const { user } = renderPlayer();

      await user.keyboard("{Escape}");

      expect(bar()).toHaveAttribute("data-state", "hidden");
    });

    it("keeps the controls up while they hold keyboard focus", async () => {
      const { user } = renderPlayer();
      await emit("playing");
      await user.keyboard("{Escape}");

      await user.tab();

      expect(bar()).toHaveAttribute("data-state", "shown");
    });
  });

  describe("keyboard scope", () => {
    it("leaves Space to a button focused outside the player", async () => {
      const { user } = renderPlayer();
      await emit("playing");
      render(<button type="button">Elsewhere</button>);
      screen.getByRole("button", { name: "Elsewhere" }).focus();
      pause.mockClear();

      await user.keyboard("[Space]");

      expect(pause).not.toHaveBeenCalled();
    });

    it("leaves shortcuts held with Ctrl to the browser", async () => {
      const { user } = renderPlayer();

      await user.keyboard("{Control>}f{/Control}");
      await user.keyboard("{Control>}{ArrowRight}{/Control}");

      expect(seek).not.toHaveBeenCalled();
    });

    it("ignores the keyboard while its tab is in the background", async () => {
      render(<SessionPlayer logs="test-logs" visible={false} />);
      const user = userEvent.setup();

      await user.keyboard("{ArrowRight}");

      expect(seek).not.toHaveBeenCalled();
    });
  });

  describe("shortcuts help", () => {
    it("toggles the shortcuts popover", async () => {
      const { user } = renderPlayer();

      expect(screen.queryByText("Keyboard Shortcuts")).not.toBeInTheDocument();

      await user.click(
        screen.getByRole("button", { name: "Keyboard shortcuts" }),
      );

      expect(screen.getByText("Keyboard Shortcuts")).toBeInTheDocument();
    });
  });
});
