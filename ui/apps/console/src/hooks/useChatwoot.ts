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
import { attempt } from "@/utils/failure";

export type ChatwootStatus =
  "non-cloud" | "unavailable" | "no-subscription" | "loading" | "ready";

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

    const identified = attempt(() => {
      window.$chatwoot?.setUser(userId, {
        email: userEmail ?? undefined,
        name: userName ?? undefined,
        identifier_hash: identifier,
      });
    });

    if (identified) lastIdentityRef.current = key;
  }, [widgetReady, userId, userEmail, userName, tenant, identifier]);

  useEffect(() => {
    if (!widgetReady || !tenant || !namespaceName) return;

    let fired = false;
    const onMessage = () => {
      if (fired) return;
      fired = true;
      attempt(() => {
        window.$chatwoot?.setConversationCustomAttributes({
          namespace: namespaceName,
          tenant,
          domain: window.location.hostname,
        });
      });
    };

    window.addEventListener("chatwoot:on-message", onMessage);
    return () => window.removeEventListener("chatwoot:on-message", onMessage);
  }, [widgetReady, tenant, namespaceName]);

  const openWidget = useCallback(() => {
    if (!widgetReady) return;
    attempt(() => window.$chatwoot?.toggle("open"));
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
