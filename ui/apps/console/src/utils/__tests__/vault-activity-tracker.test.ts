import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { start, stop, reset } from "../vault-activity-tracker";

beforeEach(() => {
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
});

afterEach(() => {
  stop();
  vi.useRealTimers();
});

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
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      delete (document as any).visibilityState;
    // eslint-disable-next-line no-empty -- the property is not configurable in every jsdom version
    } catch {
    }
  }
}

const IDLE_MS = 5_000;
const GRACE_MS = 2_000;

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

  it("re-arms idle timer on activity so it does not fire early", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    vi.advanceTimersByTime(IDLE_MS - 100);
    window.dispatchEvent(new Event("mousemove"));

    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });

  it("never fires when idleTimeoutMs is 0", () => {
    const onIdle = vi.fn();
    start({ idleTimeoutMs: 0, lockOnHidden: false, hiddenGraceMs: 0, onIdle });

    vi.advanceTimersByTime(1_000_000);
    expect(onIdle).not.toHaveBeenCalled();
  });
});

describe("first-activity throttle", () => {
  it("first activity immediately after start() triggers exactly one re-arm", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    window.dispatchEvent(new Event("mousemove"));

    vi.advanceTimersByTime(IDLE_MS - 1);
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1);
    expect(onIdle).toHaveBeenCalledOnce();
  });
});

describe("throttle", () => {
  it("second event within throttle window does NOT re-arm; idle fires at t=0-based deadline", () => {
    const onIdle = vi.fn();
    start({
      idleTimeoutMs: 60_000,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle,
    });

    vi.setSystemTime(0);
    window.dispatchEvent(new Event("mousemove"));

    vi.advanceTimersByTime(500);
    window.dispatchEvent(new Event("mousedown"));

    vi.advanceTimersByTime(60_000 - 500 - 1); // total elapsed: 59_999ms
    expect(onIdle).not.toHaveBeenCalled();

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

    vi.setSystemTime(0);
    window.dispatchEvent(new Event("mousemove"));

    vi.advanceTimersByTime(1_500);
    window.dispatchEvent(new Event("keydown"));

    vi.advanceTimersByTime(60_000 - 1_500 - 1); // total elapsed: 59_999ms
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1_500 - 1); // total elapsed: 61_498ms (= 60_000 + 1_498)
    expect(onIdle).not.toHaveBeenCalled();

    vi.advanceTimersByTime(2); // total elapsed: 61_500ms
    expect(onIdle).toHaveBeenCalledOnce();
  });
});

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

    vi.advanceTimersByTime(IDLE_MS / 2);

    start({
      idleTimeoutMs: IDLE_MS,
      lockOnHidden: false,
      hiddenGraceMs: 0,
      onIdle: onIdle2,
    });

    vi.advanceTimersByTime(IDLE_MS / 2 + 1);

    expect(onIdle1).not.toHaveBeenCalled();

    vi.advanceTimersByTime(IDLE_MS / 2);
    expect(onIdle2).toHaveBeenCalledOnce();
  });
});

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
