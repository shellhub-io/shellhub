export {};

declare global {
  interface ChatwootUserAttributes {
    email?: string;
    name?: string;
    avatar_url?: string;
    identifier_hash?: string;
  }

  interface ChatwootSettings {
    locale?: string;
    position?: "left" | "right";
    hideMessageBubble?: boolean;
    type?: "standard" | "expanded_bubble";
    launcherTitle?: string;
  }

  interface ChatwootSDK {
    setUser: (id: string, attrs: ChatwootUserAttributes) => void;
    setCustomAttributes: (attrs: Record<string, unknown>) => void;
    setConversationCustomAttributes: (attrs: Record<string, unknown>) => void;
    deleteCustomAttribute: (key: string) => void;
    toggle: (state?: "open" | "close") => void;
    reset: () => void;
  }

  interface Window {
    chatwootSDK?: { run: (cfg: { websiteToken: string; baseUrl: string }) => void };
    chatwootSettings?: ChatwootSettings;
    $chatwoot?: ChatwootSDK;
  }
}
