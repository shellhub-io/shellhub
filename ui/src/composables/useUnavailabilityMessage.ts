import { envVariables } from "@/envVariables";

interface UnavailabilityMessage {
  title: string;
  description: string;
  actionText: string;
  icon: string;
  supportLink?: {
    text: string;
    url: string;
  };
}

export function useUnavailabilityMessage(): UnavailabilityMessage {
  if (envVariables.isCloud) {
    return {
      title: "System Temporarily Unavailable",
      description: "ShellHub is temporarily down, but we're working hard to restore service. We'll be back online soon!",
      actionText: "Retry Connection",
      icon: "mdi-cloud-alert",
      supportLink: {
        text: "View Status Page",
        url: "https://status.shellhub.io/",
      },
    };
  }

  if (envVariables.isEnterprise) {
    return {
      title: "System Temporarily Unavailable",
      description: `ShellHub is currently unavailable.
          We are working on it and will be back shortly. We apologize for any inconvenience.`,
      actionText: "Try Again",
      icon: "mdi-tools",
      supportLink: {
        text: "Report Issue",
        url: "https://github.com/shellhub-io/shellhub/issues",
      },
    };
  }

  return {
    title: "Unable to Connect to API",
    description: "We're having trouble connecting to your ShellHub server. Please check that your API is running and accessible.",
    actionText: "Retry Connection",
    icon: "mdi-server-network-off",
    supportLink: {
      text: "Troubleshooting Guide",
      url: "https://docs.shellhub.io/troubleshooting/",
    },
  };
}
