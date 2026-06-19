import { useState } from "react";
import { LifebuoyIcon } from "@heroicons/react/24/outline";
import { useChatwootContext } from "@/hooks/useChatwoot";
import SupportPaywallDialog from "./SupportPaywallDialog";
import { Spinner, IconButton } from "@shellhub/design-system/primitives";

const GITHUB_ISSUE_URL = "https://github.com/shellhub-io/shellhub/issues/new";

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
        className="inline-flex items-center justify-center w-8 h-8 rounded-lg transition-all duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary/50 focus-visible:outline-offset-2 text-text-secondary hover:bg-hover-medium hover:text-text-primary border border-transparent hover:border-border"
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
      <IconButton
        onClick={handleClick}
        aria-disabled={isLoading || undefined}
        aria-label={ariaLabel}
        className={isLoading ? "cursor-wait" : undefined}
      >
        {isLoading ? (
          <Spinner size="sm" tone="subtle" className="block" />
        ) : (
          <LifebuoyIcon className="w-[18px] h-[18px]" aria-hidden="true" />
        )}
      </IconButton>
      <SupportPaywallDialog
        open={paywallOpen && status === "no-subscription"}
        onClose={() => setPaywallOpen(false)}
      />
    </>
  );
}
