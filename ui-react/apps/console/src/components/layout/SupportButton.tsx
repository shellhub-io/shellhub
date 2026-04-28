import { useState } from "react";
import { LifebuoyIcon } from "@heroicons/react/24/outline";
import { useChatwootContext } from "@/hooks/useChatwoot";
import SupportPaywallDialog from "./SupportPaywallDialog";

const GITHUB_ISSUE_URL = "https://github.com/shellhub-io/shellhub/issues/new";

const ICON_BUTTON_BASE =
  "inline-flex items-center justify-center w-8 h-8 rounded-lg transition-all duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary/50 focus-visible:outline-offset-2";

const ICON_BUTTON_ENABLED =
  "text-text-secondary hover:bg-hover-medium hover:text-text-primary border border-transparent hover:border-border";

export default function SupportButton() {
  const { status, openWidget } = useChatwootContext();
  const [paywallOpen, setPaywallOpen] = useState(false);

  if (status === "unavailable") return null;

  if (status === "non-cloud") {
    return (
      <a
        href={GITHUB_ISSUE_URL}
        target="_blank"
        rel="noopener noreferrer"
        aria-label="Report an issue on GitHub"
        className={`${ICON_BUTTON_BASE} ${ICON_BUTTON_ENABLED}`}
      >
        <LifebuoyIcon className="w-[18px] h-[18px]" aria-hidden="true" />
      </a>
    );
  }

  const isLoading = status === "loading";

  const handleClick = () => {
    if (isLoading) return;
    if (status === "no-subscription") {
      setPaywallOpen(true);
      return;
    }
    openWidget();
  };

  const ariaLabel = isLoading
    ? "Loading support chat…"
    : status === "no-subscription"
      ? "Open support — paid plan required"
      : "Open support chat";

  return (
    <>
      <button
        type="button"
        onClick={handleClick}
        aria-disabled={isLoading || undefined}
        aria-label={ariaLabel}
        className={`${ICON_BUTTON_BASE} ${
          isLoading ? "text-text-muted cursor-wait" : ICON_BUTTON_ENABLED
        }`}
      >
        {isLoading ? (
          <span
            aria-hidden="true"
            className="block w-3.5 h-3.5 border-2 border-text-muted/30 border-t-text-secondary rounded-full animate-spin"
          />
        ) : (
          <LifebuoyIcon className="w-[18px] h-[18px]" aria-hidden="true" />
        )}
      </button>
      <SupportPaywallDialog
        open={paywallOpen && status === "no-subscription"}
        onClose={() => setPaywallOpen(false)}
      />
    </>
  );
}
