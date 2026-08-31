import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { getConfig, isCloud } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useSupportIdentifier } from "@/hooks/useSupportIdentifier";
import { hasActiveSubscription } from "@/utils/billing";
import {
  falseSnapshot,
  getBootstrapFailedSnapshot,
  getWidgetReadySnapshot,
  injectChatwootScript,
  subscribeChatwootState,
} from "@/hooks/chatwootRuntime";

/**
 * Why support chat is or is not available. The three unavailable reasons are kept apart because
 * they call for different UI: nothing at all, a retry, or an upgrade.
 */
export type ChatwootStatus =
  "non-cloud" | "unavailable" | "no-subscription" | "loading" | "ready";

/**
 * What a consumer of the support widget gets: whether it can be opened, and how.
 */
export interface ChatwootHandle {
  status: ChatwootStatus;
  openWidget: () => void;
}

/**
 * Carries the widget handle down the tree. The script may be loaded only once per page, so the
 * provider owns it and everything else reads it from here.
 */
export const ChatwootContext = createContext<ChatwootHandle | null>(null);

/**
 * The widget handle from context. Throws outside a ChatwootProvider rather than returning null,
 * so the mistake surfaces where it was made.
 */
export function useChatwootContext(): ChatwootHandle {
  const ctx = useContext(ChatwootContext);
  if (!ctx) {
    throw new Error(
      "useChatwootContext must be used within a ChatwootProvider",
    );
  }
  return ctx;
}

/**
 * Loads and manages the support widget. For the provider only — everything else should use
 * useChatwootContext, since calling this twice would inject the script twice.
 */
export function useChatwoot(): ChatwootHandle {
  const config = getConfig();

  const userId = useAuthStore((s) => s.userId);
  const userEmail = useAuthStore((s) => s.email);
  const userName = useAuthStore((s) => s.name);
  const tenant = useAuthStore((s) => s.tenant);

  const { namespace } = useNamespace(tenant ?? "");
  const namespaceName = namespace?.name ?? "";
  const hasActiveBilling = hasActiveSubscription(namespace?.billing);

  const isCloudEdition = isCloud();
  const hasCloudConfig =
    !!config.chatwootWebsiteToken && !!config.chatwootBaseUrl;

  const { identifier, isError: identifierError } = useSupportIdentifier(
    tenant,
    isCloudEdition && hasCloudConfig && hasActiveBilling,
  );

  const widgetReady = useSyncExternalStore(
    subscribeChatwootState,
    getWidgetReadySnapshot,
    falseSnapshot,
  );
  const widgetFailed = useSyncExternalStore(
    subscribeChatwootState,
    getBootstrapFailedSnapshot,
    falseSnapshot,
  );

  const lastIdentityRef = useRef<string | null>(null);

  useEffect(() => {
    if (!isCloudEdition || !hasCloudConfig) return;
    if (!hasActiveBilling || !identifier || !userId) return;
    injectChatwootScript({
      websiteToken: config.chatwootWebsiteToken,
      baseUrl: config.chatwootBaseUrl,
    });
  }, [
    isCloudEdition,
    hasCloudConfig,
    hasActiveBilling,
    identifier,
    userId,
    config.chatwootBaseUrl,
    config.chatwootWebsiteToken,
  ]);

  useEffect(() => {
    if (!widgetReady || !userId || !identifier) return;
    const key = [
      userId,
      userEmail ?? "",
      userName ?? "",
      tenant ?? "",
      identifier,
    ].join("|");
    if (lastIdentityRef.current === key) return;

    try {
      window.$chatwoot?.setUser(userId, {
        email: userEmail ?? undefined,
        name: userName ?? undefined,
        identifier_hash: identifier,
      });
      lastIdentityRef.current = key;
    // eslint-disable-next-line no-empty -- the widget was reset between effect setup and this call; the next change retries
    } catch {
    }
  }, [widgetReady, userId, userEmail, userName, tenant, identifier]);

  useEffect(() => {
    if (!widgetReady || !tenant || !namespaceName) return;

    let fired = false;
    const onMessage = () => {
      if (fired) return;
      fired = true;
      try {
        window.$chatwoot?.setConversationCustomAttributes({
          namespace: namespaceName,
          tenant,
          domain: window.location.hostname,
        });
      // eslint-disable-next-line no-empty -- older Chatwoot builds do not expose this API
      } catch {
      }
    };

    window.addEventListener("chatwoot:on-message", onMessage);
    return () => window.removeEventListener("chatwoot:on-message", onMessage);
  }, [widgetReady, tenant, namespaceName]);

  const openWidget = useCallback(() => {
    if (!widgetReady) return;
    try {
      window.$chatwoot?.toggle("open");
    // eslint-disable-next-line no-empty -- the widget is not attached yet, so opening it is a no-op
    } catch {
    }
  }, [widgetReady]);

  let status: ChatwootStatus;
  if (!isCloudEdition) {
    status = "non-cloud";
  } else if (!hasCloudConfig) {
    status = "unavailable";
  } else if (widgetFailed || identifierError) {
    status = "unavailable";
  } else if (!namespace) {
    status = "loading";
  } else if (!hasActiveBilling) {
    status = "no-subscription";
  } else if (!widgetReady || !identifier) {
    status = "loading";
  } else {
    status = "ready";
  }

  return { status, openWidget };
}
