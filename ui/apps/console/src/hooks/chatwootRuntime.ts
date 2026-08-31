import { attempt } from "@/utils/failure";
/**
 * Module-level lifecycle for the Chatwoot SDK. Lives outside the React tree
 * so `authStore.logout()` can tear it down without going through a hook, and
 * so a single SPA session shares one widget across mounts/unmounts.
 *
 * The script-injection latch (`scriptInjected`) prevents StrictMode and
 * dependency churn from stacking multiple <script> tags. The bootstrap
 * watchdog flips `bootstrapFailed` when `chatwoot:ready` never fires after a
 * successful `script.onload` (invalid token, content blocker, SDK throws),
 * so the UI can surface "unavailable" instead of spinning forever.
 */

const SCRIPT_ID = "shellhub-chatwoot-sdk";
const BOOTSTRAP_TIMEOUT_MS = 15_000;

// CSS selectors for Chatwoot's own injected DOM (the floating launcher,
// chat bubble, iframe). Removed on tear-down so the widget doesn't visually
// linger after logout — the SDK attaches these directly to <body>, outside
// the React tree, so unmounting AppLayout doesn't reach them.
const SDK_DOM_SELECTORS =
  '.woot-widget-holder, .woot--bubble-holder, #cw-bubble-holder, iframe[src*="chatwoot"]';

let scriptInjected = false;
let bootstrapFailed = false;
let bootstrapTimer: ReturnType<typeof setTimeout> | null = null;
let readyListener: (() => void) | null = null;

const listeners = new Set<() => void>();

function notify(): void {
  for (const cb of listeners) cb();
}

function clearWatchdog(): void {
  if (bootstrapTimer !== null) {
    clearTimeout(bootstrapTimer);
    bootstrapTimer = null;
  }
}

function detachReadyListener(): void {
  if (readyListener && typeof window !== "undefined") {
    window.removeEventListener("chatwoot:ready", readyListener);
  }
  readyListener = null;
}

/**
 * Subscribes to changes in the widget's readiness, returning the unsubscribe. The store half of
 * useSyncExternalStore: Chatwoot loads asynchronously and outside React, so its state cannot
 * be React state.
 */
export function subscribeChatwootState(callback: () => void): () => void {
  listeners.add(callback);
  return () => {
    listeners.delete(callback);
  };
}

/**
 * Whether the widget is loaded and callable. Read from the global the script installs, so it is
 * correct even if the script loaded before React did.
 */
export function getWidgetReadySnapshot(): boolean {
  return typeof window !== "undefined" && !!window.$chatwoot;
}

/**
 * Whether the widget failed to load. Distinct from not-ready: a failure is final, and the UI
 * stops offering support rather than waiting.
 */
export function getBootstrapFailedSnapshot(): boolean {
  return bootstrapFailed;
}

/**
 * The server snapshot for useSyncExternalStore. There is no widget outside the browser, so it is
 * always false — a constant function rather than an inline literal, which would change identity
 * on every render and loop.
 */
export function falseSnapshot(): boolean {
  return false;
}

interface InjectArgs {
  websiteToken: string;
  baseUrl: string;
}

/**
 * Inject the Chatwoot SDK <script> tag once per session. No-op if the widget
 * is already attached or another mount is mid-injection. Schedules a
 * watchdog that flips `bootstrapFailed` if the widget never reports ready.
 */
export function injectChatwootScript({
  websiteToken,
  baseUrl,
}: InjectArgs): void {
  if (typeof window === "undefined") return;
  if (getWidgetReadySnapshot()) return;

  if (scriptInjected && !document.getElementById(SCRIPT_ID)) {
    scriptInjected = false;
  }

  if (scriptInjected || document.getElementById(SCRIPT_ID)) return;

  bootstrapFailed = false;
  scriptInjected = true;
  const cleanBaseUrl = baseUrl.replace(/\/+$/, "");

  window.chatwootSettings = {
    locale: "en",
    position: "right",
    hideMessageBubble: true,
    type: "standard",
  };

  detachReadyListener();
  readyListener = () => {
    clearWatchdog();
    bootstrapFailed = false;
    detachReadyListener();
    notify();
  };
  window.addEventListener("chatwoot:ready", readyListener, { once: true });

  const script = document.createElement("script");
  script.id = SCRIPT_ID;
  script.src = `${cleanBaseUrl}/packs/js/sdk.js`;
  script.async = true;
  script.onload = () => {
    if (!window.chatwootSDK) {
      tearDownChatwoot("bootstrap-timeout");
      return;
    }
    try {
      window.chatwootSDK.run({ websiteToken, baseUrl: cleanBaseUrl });
    } catch {
      tearDownChatwoot("bootstrap-timeout");
      return;
    }
    bootstrapTimer = setTimeout(() => {
      bootstrapTimer = null;
      if (!getWidgetReadySnapshot()) {
        tearDownChatwoot("bootstrap-timeout");
      }
    }, BOOTSTRAP_TIMEOUT_MS);
  };
  script.onerror = () => {
    tearDownChatwoot("bootstrap-timeout");
  };
  document.body.appendChild(script);
}

/**
 * Remove the SDK from the page entirely. Called on logout (so the next user
 * gets a fresh widget) and from the bootstrap watchdog (so a wedged widget
 * surfaces as "unavailable" rather than spinning forever).
 *
 * Must run synchronously while AppLayout is still mounted — once the route
 * changes to /login, ChatwootProvider unmounts and we lose the chance to
 * call $chatwoot.reset() against the live iframe.
 */
export function tearDownChatwoot(
  reason: "logout" | "bootstrap-timeout" = "logout",
): void {
  if (typeof window === "undefined") return;

  clearWatchdog();
  detachReadyListener();

  attempt(() => {
    window.$chatwoot?.toggle("close");
    window.$chatwoot?.reset();
  });

  document.getElementById(SCRIPT_ID)?.remove();
  document.querySelectorAll(SDK_DOM_SELECTORS).forEach((node) => node.remove());

  delete window.$chatwoot;
  delete window.chatwootSDK;
  delete window.chatwootSettings;
  scriptInjected = false;
  bootstrapFailed = reason === "bootstrap-timeout";
  notify();
}
