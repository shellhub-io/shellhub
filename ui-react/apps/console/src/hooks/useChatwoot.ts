import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { getConfig } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useSupportIdentifier } from "@/hooks/useSupportIdentifier";
import {
  falseSnapshot,
  getBootstrapFailedSnapshot,
  getWidgetReadySnapshot,
  injectChatwootScript,
  subscribeChatwootState,
} from "@/hooks/chatwootRuntime";
import { readNamespaceBilling } from "@/types/billing";

export type ChatwootStatus =
  | "non-cloud"
  | "unavailable"
  | "no-subscription"
  | "loading"
  | "ready";

export interface ChatwootHandle {
  status: ChatwootStatus;
  openWidget: () => void;
}

export const ChatwootContext = createContext<ChatwootHandle | null>(null);

export function useChatwootContext(): ChatwootHandle {
  const ctx = useContext(ChatwootContext);
  if (!ctx) {
    throw new Error(
      "useChatwootContext must be used within a ChatwootProvider",
    );
  }
  return ctx;
}

export function useChatwoot(): ChatwootHandle {
  const config = getConfig();

  const userId = useAuthStore((s) => s.userId);
  const userEmail = useAuthStore((s) => s.email);
  const userName = useAuthStore((s) => s.name);
  const tenant = useAuthStore((s) => s.tenant);

  const { namespace } = useNamespace(tenant ?? "");
  const namespaceName = namespace?.name ?? "";
  const billing = readNamespaceBilling(namespace?.billing);
  const hasActiveBilling = billing?.active === true;

  const isCloud = config.cloud;
  const hasCloudConfig =
    !!config.chatwootWebsiteToken && !!config.chatwootBaseUrl;

  const { identifier, isError: identifierError } = useSupportIdentifier(
    tenant,
    isCloud && hasCloudConfig && hasActiveBilling,
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

  // Inject the SDK script once prerequisites are met. The runtime helper is
  // idempotent across StrictMode double-effects and concurrent mounts.
  useEffect(() => {
    if (!isCloud || !hasCloudConfig) return;
    if (!hasActiveBilling || !identifier || !userId) return;
    injectChatwootScript({
      websiteToken: config.chatwootWebsiteToken,
      baseUrl: config.chatwootBaseUrl,
    });
  }, [
    isCloud,
    hasCloudConfig,
    hasActiveBilling,
    identifier,
    userId,
    config.chatwootBaseUrl,
    config.chatwootWebsiteToken,
  ]);

  // Identify the user (or re-identify on profile/tenant/identifier change).
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
    } catch {
      // Widget reset between effect setup and call — next change retries.
    }
  }, [widgetReady, userId, userEmail, userName, tenant, identifier]);

  // Set conversation attributes on the next message. Gated on namespaceName
  // so the first message after a tenant switch isn't tagged with an empty
  // namespace while the new namespace fetch is still in flight.
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
      } catch {
        // Ignore — Chatwoot may not expose the API in older builds.
      }
    };

    window.addEventListener("chatwoot:on-message", onMessage);
    return () => window.removeEventListener("chatwoot:on-message", onMessage);
  }, [widgetReady, tenant, namespaceName]);

  const openWidget = useCallback(() => {
    if (!widgetReady) return;
    try {
      window.$chatwoot?.toggle("open");
    } catch {
      // Widget not yet attached — no-op.
    }
  }, [widgetReady]);

  let status: ChatwootStatus;
  if (!isCloud) {
    status = "non-cloud";
  } else if (!hasCloudConfig) {
    status = "unavailable";
  } else if (widgetFailed || identifierError) {
    // Operator misconfiguration (missing identity key) or definitive SDK
    // bootstrap failure — hide the button rather than spin forever.
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
