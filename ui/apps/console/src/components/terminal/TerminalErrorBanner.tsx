import { Link } from "react-router-dom";
import { ExclamationCircleIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import { useTerminalStore } from "@/stores/terminalStore";
import type { TerminalError } from "./terminalErrors";

interface TerminalErrorBannerProps {
  error: TerminalError;
  sessionId: string;
}

export default function TerminalErrorBanner({
  error,
  sessionId,
}: TerminalErrorBannerProps) {
  return (
    <div
      role="alert"
      className="bg-accent-red/[0.08] border-b border-accent-red/20 px-5 py-3.5 flex items-start gap-3 animate-slide-down"
    >
      <ExclamationCircleIcon
        className="w-4 h-4 text-accent-red shrink-0 mt-0.5"
        strokeWidth={1.5}
      />
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2 mb-1">
          <span className="text-sm font-semibold text-text-primary">
            {error.title}
          </span>
          <span className="text-sm text-text-muted">{error.message}</span>
        </div>
        {error.hints.length > 0 && (
          <p className="text-sm text-text-secondary leading-relaxed mb-1.5">
            {error.hints.join(" ")}
          </p>
        )}
        {(error.links.length > 0 || error.reconnect) && (
          <div className="flex items-center gap-3">
            {error.links.map((link) => (
              <Link
                key={link.to}
                to={link.to}
                onClick={() => useTerminalStore.getState().close(sessionId)}
                className="text-sm text-primary hover:text-primary-600 font-medium transition-colors"
              >
                {link.label}
              </Link>
            ))}
            {error.links.length > 0 && error.reconnect && (
              <span className="w-px h-3.5 bg-border-light" />
            )}
            {error.reconnect && (
              <Button
                size="sm"
                onClick={() => {
                  useTerminalStore.getState().closeAndReconnect(sessionId);
                }}
              >
                Retry
              </Button>
            )}
          </div>
        )}
      </div>
      <IconButton
        size="sm"
        aria-label="Dismiss"
        onClick={() => useTerminalStore.getState().close(sessionId)}
      >
        <XMarkIcon className="w-3.5 h-3.5" />
      </IconButton>
    </div>
  );
}
