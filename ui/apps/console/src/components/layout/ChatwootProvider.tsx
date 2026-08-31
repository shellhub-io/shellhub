import type { ReactNode } from "react";
import { ChatwootContext, useChatwoot } from "@/hooks/useChatwoot";

/**
 * Loads the support widget once and shares the handle below it. Mounted at the top of the app,
 * since the script may only be injected once per page.
 */
export default function ChatwootProvider({ children }: { children: ReactNode }) {
  const handle = useChatwoot();
  return (
    <ChatwootContext.Provider value={handle}>
      {children}
    </ChatwootContext.Provider>
  );
}
