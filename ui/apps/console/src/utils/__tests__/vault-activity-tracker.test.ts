import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { start, stop, reset } from "../vault-activity-tracker";

// ---------------------------------------------------------------------------
// Fake-timer helpers
// ---------------------------------------------------------------------------

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
});

afterEach(() => {
  stop();
  vi.useRealTimers();
});

// ---------------------------------------------------------------------------
// Helpers for visibilityState override
// ---------------------------------------------------------------------------

let originalVisibilityDescriptor: PropertyDescriptor | undefined;

function setVisibility(state: "visible" | "hidden") {
  if (!originalVisibilityDescriptor) {
    originalVisibilityDescriptor =
      Object.getOwnPropertyDescriptor(Document.prototype, "visibilityState") ??
      Object.getOwnPropertyDescriptor(document, "visibilityState");
  }
  Object.defineProperty(document, "visibilityState", {
    configurable: true,
    get: () => state,
  });
  document.dispatchEvent(new Event("visibilitychange"));
}

function restoreVisibility() {
  if (originalVisibilityDescriptor) {
    Object.defineProperty(
      document,
      "visibilityState",
      originalVisibilityDescriptor,
    );
    originalVisibilityDescriptor = undefined;
  } else {
    // Delete the override so the prototype chain is visible again
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      delete (document as any).visibilityState;
    } catch {
      // ignore
    }
  }
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const IDLE_MS = 5_000;
const GRACE_MS = 2_000;

// ---------------------------------------------------------------------------
// 1. onIdle fires after idleTimeoutMs with no activity
// ---------------------------------------------------------------------------

describe("idle timer", () => {
  it("fires onIdle after idleTimeoutMs elapses with no activity", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  // ---------------------------------------------------------------------------
  // 2. activity event re-arms idle timer (no premature fire)
  // ---------------------------------------------------------------------------

  it("re-arms idle timer on activity so it does not fire early", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    // Advance close to the deadline, then fire activity to re-arm
    vi.advanceTimersByTime(IDLE_MS - 100);
    window.dispatchEvent(new Event("mousemove"));

    // Advance by another IDLE_MS - 1 (total way past original deadline)
    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    // Now finish the re-armed timer
    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  // ---------------------------------------------------------------------------
  // 3. idleTimeoutMs:0 never fires
  // ---------------------------------------------------------------------------

  it("never fires when idleTimeoutMs is 0", () => {
    const onIdle = vi.fn();
    start({ idleTimeoutMs: 0, lockOnHidden: false, hiddenGraceMs: 0, onIdle });

    vi.advanceTimersByTime(1_000_000);
    expect(onIdle).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// 4. First activity immediately after start() re-arms exactly once
// ---------------------------------------------------------------------------

describe("first-activity throttle", () => {
  it("first activity immediately after start() triggers exactly one re-arm", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    // Fire activity at t=0 (well within throttle window relative to stamp=0)
    window.dispatchEvent(new Event("mousemove"));

    // Advance past idle — it was re-armed by the above event
    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });
});

// ---------------------------------------------------------------------------
// 5. Throttle: at most one reset per 1000ms
// ---------------------------------------------------------------------------

describe("throttle", () => {
  it("second event within throttle window does NOT re-arm; idle fires at t=0-based deadline", () => {
    const onIdle = vi.fn();
    // idleTimeoutMs=60_000 so the deadline is clearly separated from the throttle window
    start({
      idleTimeoutMs: 60_000,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    // t=0: first event — passes throttle (stamp was 0), re-arms idle timer to fire at t=60_000
    vi.setSystemTime(0);
    window.dispatchEvent(new Event("mousemove"));

    // t=500ms: second event — within 1000ms throttle window, should NOT re-arm the timer.
    // If it DID re-arm, the new deadline would be t=500+60_000=60_500ms, meaning onIdle
    // would NOT fire at exactly t=60_000ms below — which is what would happen if the
    // throttle guard were removed.
    vi.advanceTimersByTime(500);
    window.dispatchEvent(new Event("mousedown"));

    // Advance to 1ms before the t=0-based deadline: timer must not have fired yet.
    vi.advanceTimersByTime(60_000 - 500 - 1); // total elapsed: 59_999ms
    expect(onIdle).not.toHaveBeenCalled();

    // Advance the final 1ms to exactly t=60_000ms.
    // The idle MUST fire here because the timer was last armed at t=0 (the second event
    // at t=500 was throttled and did not re-arm it).
    // If the throttle guard is broken the timer would have been re-armed at t=500ms and
    // would still have 500ms remaining — so onIdle would NOT fire, making this assertion fail.
    vi.advanceTimersByTime(1); // total elapsed: 60_000ms
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it("event outside throttle window DOES re-arm; idle fires from that event's deadline", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: 60_000,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    // t=0: first event — re-arms idle timer to fire at t=60_000
    vi.setSystemTime(0);
    window.dispatchEvent(new Event("mousemove"));

    // t=1_500ms: third event — >1s after last re-arm at t=0, passes throttle, re-arms to t=61_500
    vi.advanceTimersByTime(1_500);
    window.dispatchEvent(new Event("keydown"));

    // At t=60_000ms the original deadline passes; onIdle must NOT fire because the timer
    // was re-armed at t=1_500ms.
    vi.advanceTimersByTime(60_000 - 1_500 - 1); // total elapsed: 59_999ms
    expect(onIdle).not.toHaveBeenCalled();

    // Still not fired at just before t=61_500ms
    vi.advanceTimersByTime(1_500 - 1); // total elapsed: 61_498ms (= 60_000 + 1_498)
    expect(onIdle).not.toHaveBeenCalled();

    // Now fires at exactly t=61_500ms
    vi.advanceTimersByTime(2); // total elapsed: 61_500ms
    expect(onIdle).toHaveBeenCalledOnce();
  });
});

// ---------------------------------------------------------------------------
// 6. stop() cancels timers, removes listeners, and is idempotent
// ---------------------------------------------------------------------------

describe("stop()", () => {
  it("cancels the idle timer so onIdle never fires after stop()", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    vi.advanceTimersByTime(IDLE_MS / 2);
    stop();
    vi.advanceTimersByTime(IDLE_MS);

    expect(onIdle).not.toHaveBeenCalled();
  });

  it("removes activity listeners so events no longer re-arm after stop()", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });
    stop();

    // Events dispatched after stop must not cause any effect
    window.dispatchEvent(new Event("mousemove"));
    window.dispatchEvent(new Event("keydown"));
    vi.advanceTimersByTime(IDLE_MS * 2);

    expect(onIdle).not.toHaveBeenCalled();
  });

  it("is idempotent — calling stop() multiple times does not throw", () => {
    expect(() => {
      stop();
      stop();
      stop();
    }).not.toThrow();
  });

  it("is safe to call stop() before start()", () => {
    expect(() => stop()).not.toThrow();
  });
});

// ---------------------------------------------------------------------------
// 7. lockOnHidden: hidden fires after grace period; visible cancels it
// ---------------------------------------------------------------------------

describe("lockOnHidden", () => {
  afterEach(() => {
    restoreVisibility();
  });

  it("fires onIdle after hiddenGraceMs when document becomes hidden", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: 0,
      lockOnHidden: true,
      hiddenGraceMs: GRACE_MS,
      onIdle,
    });

    setVisibility("hidden");

    vi.advanceTimersByTime(GRACE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it("cancels the grace timer when document becomes visible before grace elapses", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: 0,
      lockOnHidden: true,
      hiddenGraceMs: GRACE_MS,
      onIdle,
    });

    setVisibility("hidden");
    vi.advanceTimersByTime(GRACE_MS / 2);

    setVisibility("visible");
    vi.advanceTimersByTime(GRACE_MS); // advance well past original grace deadline

    expect(onIdle).not.toHaveBeenCalled();
  });

  it("does not attach visibilitychange listener when lockOnHidden is false", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: 0,
      lockOnHidden: false,
      hiddenGraceMs: GRACE_MS,
      onIdle,
    });

    setVisibility("hidden");
    vi.advanceTimersByTime(GRACE_MS * 2);

    expect(onIdle).not.toHaveBeenCalled();
  });
});

// ---------------------------------------------------------------------------
// 8. start() restarts cleanly (calls stop() first)
// ---------------------------------------------------------------------------

describe("start() restart", () => {
  it("calling start() a second time cancels the first session and starts fresh", () => {
    const onIdle1 = vi.fn();
    const onIdle2 = vi.fn();

    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle: onIdle1,
    });

    // Advance partway through first session
    vi.advanceTimersByTime(IDLE_MS / 2);

    // Restart with a different callback
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle: onIdle2,
    });

    // Advance past where the first timer would have fired
    vi.advanceTimersByTime(IDLE_MS / 2 + 1);

    // First timer must have been cancelled
    expect(onIdle1).not.toHaveBeenCalled();

    // Advance to complete the second timer
    vi.advanceTimersByTime(IDLE_MS / 2);
    expect(onIdle2).toHaveBeenCalledOnce();
  });
});

// ---------------------------------------------------------------------------
// 9. reset() re-arms idle timer (exported for tests)
// ---------------------------------------------------------------------------

describe("reset()", () => {
  it("re-arms the idle timer when called directly", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    vi.advanceTimersByTime(IDLE_MS - 100);
    reset();

    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });
});

// ---------------------------------------------------------------------------
// 10. All ACTIVITY_EVENTS trigger re-arm
// ---------------------------------------------------------------------------

describe("ACTIVITY_EVENTS", () => {
  const EVENTS = [
    "mousemove",
    "mousedown",
    "keydown",
    "scroll",
    "touchstart",
  ] as const;

  for (const eventName of EVENTS) {
    it(`'${eventName}' event re-arms the idle timer`, () => {
      const onIdle = vi.fn();
      start({
        idleTimeoutMs: IDLE_MS,
        lockOnHidden: false,
        hiddenGraceMs: 0,
        onIdle,
      });

      vi.advanceTimersByTime(IDLE_MS - 100);
      window.dispatchEvent(new Event(eventName));

      vi.advanceTimersByTime(IDLE_MS - 1);
      expect(onIdle).not.toHaveBeenCalled();

      vi.advanceTimersByTime(1);
      expect(onIdle).toHaveBeenCalledOnce();
    });
  }
});
