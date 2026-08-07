import type { ReactNode } from "react";
import { ArrowRightEndOnRectangleIcon } from "@heroicons/react/24/outline";
import {
  Button,
  GithubIcon,
  GoogleIcon,
} from "@shellhub/design-system/primitives";
import type { AuthConnector } from "@/client";

const connectorIcons: Record<string, ReactNode> = {
  google: <GoogleIcon className="w-4 h-4" />,
  github: <GithubIcon className="w-4 h-4" />,
};

interface AuthConnectorButtonsProps {
  connectors: AuthConnector[];
  variant?: "primary" | "secondary";
}

// The login and sign-up pages offer the same providers: on Cloud they are the social buttons, on
// enterprise the instance's own directory. Starting a login is a full-page navigation, not a fetch,
// because the browser has to follow the provider's redirects.
export default function AuthConnectorButtons({
  connectors,
  variant = "secondary",
}: AuthConnectorButtonsProps) {
  return (
    <>
      {connectors.map((connector) => (
        <Button
          key={connector.id}
          variant={variant}
          fullWidth
          icon={
            connectorIcons[connector.type] ?? (
              <ArrowRightEndOnRectangleIcon className="w-4 h-4" />
            )
          }
          data-testid={`sso-connector-${connector.id}`}
          onClick={() => {
            window.location.assign(
              `/api/auth/sso/${encodeURIComponent(connector.id)}`,
            );
          }}
        >
          {`Continue with ${connector.name}`}
        </Button>
      ))}
    </>
  );
}
