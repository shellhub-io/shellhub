import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, act, fireEvent } from "@testing-library/react";
import SessionPlayer from "../SessionPlayer";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

vi.mock("asciinema-player", () => ({
  create: vi.fn(),
}));

import { create } from "asciinema-player";

const mockedCreate = vi.mocked(create);

const mockPlay = vi.fn();
const mockPause = vi.fn();
const mockSeek = vi.fn();
const mockGetCurrentTime = vi.fn().mockResolvedValue(0);
const mockGetDuration = vi.fn().mockResolvedValue(60);
const mockDispose = vi.fn();

let listeners: Record<string, () => void> = {};

const mockPlayer = {
  play: mockPlay,
  pause: mockPause,
  seek: mockSeek,
  getCurrentTime: mockGetCurrentTime,
  getDuration: mockGetDuration,
  dispose: mockDispose,
  addEventListener: vi.fn(),
  removeEventListener: vi.fn(),
};

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.useFakeTimers();
  listeners = {};
  vi.clearAllMocks();
  mockSeek.mockResolvedValue(undefined);
  mockPlayer.addEventListener.mockImplementation(
    (event: string, handler: () => void) => {
      listeners[event] = handler;
    },
  );
  mockedCreate.mockReturnValue(mockPlayer);
});

afterEach(() => {
  vi.useRealTimers();
});

/* ------------------------------------------------------------------ */
/* Tests                                                               */
/* ------------------------------------------------------------------ */

describe("SessionPlayer", () => {
  describe("setup", () => {
    it("creates the asciinema player with the provided logs", () => {
      render(<SessionPlayer logs="my-session-data" />);

      expect(mockedCreate).toHaveBeenCalledWith(
        { data: "my-session-data" },
        expect.any(HTMLElement),
        expect.objectContaining({ controls: false }),
      );
    });

    it("starts playback automatically on mount", () => {
      render(<SessionPlayer logs="test-logs" />);

      expect(mockPlay).toHaveBeenCalledTimes(1);
    });

    it("disposes the player on unmount", () => {
      const { unmount } = render(<SessionPlayer logs="test-logs" />);
      unmount();

      expect(mockDispose).toHaveBeenCalledTimes(1);
    });
  });

  describe("seekTo", () => {
    it("resumes playback after seeking when the video was playing (ArrowRight)", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      mockPlay.mockClear();
      mockPause.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: "ArrowRight" });
      });

      expect(mockPause).toHaveBeenCalledTimes(1);
      expect(mockPlay).toHaveBeenCalledTimes(1);
    });

    it("resumes playback after seeking when the video was playing (ArrowLeft)", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      mockPlay.mockClear();
      mockPause.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: "ArrowLeft" });
      });

      expect(mockPause).toHaveBeenCalledTimes(1);
      expect(mockPlay).toHaveBeenCalledTimes(1);
    });

    it("resumes playback after seeking when the video was playing (number key)", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      mockPlay.mockClear();
      mockPause.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: "5" });
      });

      expect(mockPause).toHaveBeenCalledTimes(1);
      expect(mockPlay).toHaveBeenCalledTimes(1);
    });

    it("does not resume playback after seeking when the video was paused", async () => {
      render(<SessionPlayer logs="test-logs" />);

      // isPlayingRef stays false — "playing" listener never triggered
      mockPlay.mockClear();
      mockPause.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: "ArrowRight" });
      });

      expect(mockPause).toHaveBeenCalledTimes(1);
      expect(mockPlay).not.toHaveBeenCalled();
    });

    it("seeks back/forward one frame with comma/period when paused", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        fireEvent.keyDown(window, { key: "," });
      });

      await act(async () => {
        fireEvent.keyDown(window, { key: "." });
      });

      expect(mockSeek).toHaveBeenCalledTimes(2);
    });

    it("ignores comma/period for frame stepping when playing", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      mockSeek.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: "," });
        fireEvent.keyDown(window, { key: "." });
      });

      expect(mockSeek).not.toHaveBeenCalled();
    });
  });

  describe("keyboard shortcuts", () => {
    it("pauses playback when Space is pressed while playing", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      mockPause.mockClear();
      mockPlay.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: " " });
      });

      expect(mockPause).toHaveBeenCalledTimes(1);
      expect(mockPlay).not.toHaveBeenCalled();
    });

    it("resumes playback when Space is pressed while paused", async () => {
      render(<SessionPlayer logs="test-logs" />);

      // Never trigger "playing", so isPlayingRef stays false
      mockPlay.mockClear();

      await act(async () => {
        fireEvent.keyDown(window, { key: " " });
      });

      expect(mockPlay).toHaveBeenCalledTimes(1);
      expect(mockPause).not.toHaveBeenCalled();
    });

    it("calls onClose when Escape is pressed outside fullscreen", async () => {
      const onClose = vi.fn();
      render(<SessionPlayer logs="test-logs" onClose={onClose} />);

      await act(async () => {
        fireEvent.keyDown(window, { key: "Escape" });
      });

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("resets to paused state when the ended event fires", async () => {
      render(<SessionPlayer logs="test-logs" />);

      await act(async () => {
        listeners["playing"]?.();
      });

      await act(async () => {
        listeners["ended"]?.();
      });

      // After ended, Space should resume (isPlayingRef became false)
      mockPlay.mockClear();
      await act(async () => {
        fireEvent.keyDown(window, { key: " " });
      });

      expect(mockPlay).toHaveBeenCalledTimes(1);
    });
  });
});
