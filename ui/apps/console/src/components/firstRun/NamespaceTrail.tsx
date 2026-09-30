import { Button } from "@shellhub/design-system/primitives";
import { isCommunity } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useUserInfo } from "@/hooks/useUserInfo";
import {
  CommunityInstructions,
  NamespaceCreateForm,
} from "@/components/common/CreateNamespace";
import FirstRunLayout from "./FirstRunLayout";
import { Trail, TrailStep, UpcomingDeviceSteps } from "./Trail";
import AskAdministrator from "./AskAdministrator";

/**
 * The first run for a signed-in user with no namespace. One who may create a namespace gets the
 * trail, starting there. One who has to be added to a namespace (always on community, where
 * namespaces come from the CLI, and on the other editions when the account is barred from
 * creating them) gets only that: the role they are added with decides whether the device steps
 * apply, and the dashboard settles that once they are in.
 */
export default function NamespaceTrail() {
  const name = useAuthStore((s) => s.name);
  const email = useAuthStore((s) => s.email);
  const { user, error, refetch } = useUserInfo({ enabled: !isCommunity() });
  const eyebrow = `Welcome, ${name || email || "back"}`;

  if (!isCommunity() && error) {
    return (
      <FirstRunLayout
        eyebrow={eyebrow}
        title="Something went wrong"
        lead="We couldn't load your account. Check your connection and try again."
        signedIn
        inConsole={false}
      >
        <Button onClick={() => void refetch()}>Try again</Button>
      </FirstRunLayout>
    );
  }

  if (!isCommunity() && !user) return null;

  if (isCommunity() || user?.max_namespaces === 0) {
    return (
      <FirstRunLayout
        eyebrow={eyebrow}
        title="Join a namespace"
        lead="Everything in ShellHub lives in a namespace. Once you are in one, this page takes you there."
        signedIn
        inConsole={false}
      >
        {isCommunity() ? (
          <CommunityInstructions autoEnter />
        ) : (
          <AskAdministrator />
        )}
      </FirstRunLayout>
    );
  }

  return (
    <FirstRunLayout eyebrow={eyebrow} signedIn inConsole={false}>
      <Trail>
        <TrailStep number={1} title="Create a namespace" state="active">
          <div className="space-y-3">
            <p className="text-xs text-text-secondary">
              A namespace holds your devices and the people who can reach them.
              Its name shows up in every SSH address, so keep it short.
            </p>
            <NamespaceCreateForm />
          </div>
        </TrailStep>
        <UpcomingDeviceSteps start={2} />
      </Trail>
    </FirstRunLayout>
  );
}
