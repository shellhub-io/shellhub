import type { ReactNode } from "react";
import { ChatwootContext, useChatwoot } from "@/hooks/useChatwoot";

export default function ChatwootProvider({ children }: { children: ReactNode }) {
  const handle = useChatwoot();
  return (
    <ChatwootContext.Provider value={handle}>
      {children}
    </ChatwootContext.Provider>
  );
}
