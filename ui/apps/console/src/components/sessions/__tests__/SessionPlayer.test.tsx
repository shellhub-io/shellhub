import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { Player } from "asciinema-player";
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

function renderPlayer(props: { onClose?: () => void } = {}) {
  const user = userEvent.setup();
  render(<SessionPlayer logs="test-logs" {...props} />);
  return { user };
}

const emit = (event: string) => act(async () => listeners[event]?.());

beforeEach(() => {
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
    it("shows the duration and a pause affordance once playback starts", async () => {
      renderPlayer();
      await emit("playing");

      expect(screen.getByText("00:00 / 01:00")).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Pause" })).toBeInTheDocument();
    });

    it("advances the time readout as playback progresses", async () => {
      renderPlayer();
      await emit("playing");

      currentTime = 5;

      expect(await screen.findByText("00:05 / 01:00")).toBeInTheDocument();
    });

    it("returns to a play affordance when playback ends", async () => {
      renderPlayer();
      await emit("playing");
      await emit("ended");

      expect(screen.getByRole("button", { name: "Play" })).toBeInTheDocument();
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

    it("closes the player with Escape", async () => {
      const onClose = vi.fn();
      const { user } = renderPlayer({ onClose });

      await user.keyboard("{Escape}");

      expect(onClose).toHaveBeenCalledOnce();
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

      await user.selectOptions(
        screen.getByRole("combobox", { name: "Playback speed" }),
        "2",
      );

      expect(dispose).toHaveBeenCalledTimes(1);
      expect(mockedCreate).toHaveBeenLastCalledWith(
        { data: "test-logs" },
        expect.any(HTMLElement),
        expect.objectContaining({ speed: 2, startAt: 10 }),
      );
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
