import { Button } from "@shellhub/design-system/primitives";
import { isCloud, isCommunity } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useUserInfo } from "@/hooks/useUserRequests";
import { canCreateNamespace } from "@/utils/namespaceCreation";
import { useSurveyBeforeNamespace } from "@/hooks/useOnboardingSurvey";
import {
  CommunityInstructions,
  NamespaceCreateForm,
} from "@/components/common/CreateNamespace";
import FirstRunLayout from "./FirstRunLayout";
import { Trail, TrailStep, UpcomingDeviceSteps } from "./Trail";
import { SETUP_STEP_TITLES } from "./stepTitles";
import AskAdministrator from "./AskAdministrator";
import OnboardingStep from "./survey/OnboardingStep";
import BackToSurveyButton from "./survey/BackToSurveyButton";

function CreateNamespaceTrail({ eyebrow }: { eyebrow: string }) {
  const survey = useSurveyBeforeNamespace(isCloud());
  const namespaceStep = survey.visible ? 2 : 1;

  return (
    <FirstRunLayout eyebrow={eyebrow} signedIn inConsole={false}>
      <Trail>
        {survey.visible && (
          <TrailStep
            number={1}
            title={SETUP_STEP_TITLES.survey}
            state={survey.active ? "active" : "done"}
            summary={survey.responseId ? "Thanks" : "Skipped"}
          >
            <OnboardingStep {...survey.stepProps} />
          </TrailStep>
        )}
        <TrailStep
          number={namespaceStep}
          title="Create a namespace"
          state={survey.active ? "upcoming" : "active"}
        >
          <div className="space-y-3">
            <p className="text-xs text-text-secondary">
              A namespace holds your devices and the people who can reach them.
              Its name shows up in every SSH address, so keep it short.
            </p>
            <NamespaceCreateForm />
            {survey.visible && <BackToSurveyButton onClick={survey.back} />}
          </div>
        </TrailStep>
        <UpcomingDeviceSteps start={namespaceStep + 1} />
      </Trail>
    </FirstRunLayout>
  );
}

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
  const { data: user, error, refetch } = useUserInfo();
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

  if (!canCreateNamespace(user)) {
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

  return <CreateNamespaceTrail eyebrow={eyebrow} />;
}
