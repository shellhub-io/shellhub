import { useState } from "react";
import { Button } from "@shellhub/design-system/primitives";
import { getConfig, isCloud, isCommunity } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { useUserInfo } from "@/hooks/useUserInfo";
import { useOnboardingSurvey } from "@/hooks/useOnboardingSurvey";
import {
  CommunityInstructions,
  NamespaceCreateForm,
} from "@/components/common/CreateNamespace";
import FirstRunLayout from "./FirstRunLayout";
import {
  SETUP_STEP_TITLES,
  Trail,
  TrailStep,
  UpcomingDeviceSteps,
} from "./Trail";
import AskAdministrator from "./AskAdministrator";
import OnboardingStep from "./survey/OnboardingStep";
import { emptyAnswers } from "./survey/onboardingSurvey";
import {
  mergeAnswers,
  readSavedSurvey,
  writeSavedSurvey,
  type SavedSurvey,
} from "./survey/savedSurvey";

const STEP_SURVEY = 1;
const STEP_NAMESPACE = 2;

function CreateNamespaceTrail({ eyebrow }: { eyebrow: string }) {
  const config = getConfig();
  const userId = useAuthStore((s) => s.userId) ?? "";
  const name = useAuthStore((s) => s.name);
  const email = useAuthStore((s) => s.email);

  const surveyBaseUrl = isCloud() ? config.onboardingUrl : "";
  const surveyQuery = useOnboardingSurvey(surveyBaseUrl);
  const survey = surveyQuery.data ?? null;
  const showSurvey =
    surveyBaseUrl !== "" && (surveyQuery.isPending || survey !== null);

  const [saved, setSaved] = useState<SavedSurvey | null>(() =>
    readSavedSurvey(userId),
  );
  const [chosenStep, setStep] = useState(saved ? STEP_NAMESPACE : STEP_SURVEY);
  const step = showSurvey ? chosenStep : STEP_NAMESPACE;
  const namespaceStep = showSurvey ? 2 : 1;

  return (
    <FirstRunLayout eyebrow={eyebrow} signedIn inConsole={false}>
      <Trail>
        {showSurvey && (
          <TrailStep
            number={1}
            title={SETUP_STEP_TITLES.survey}
            state={step === STEP_SURVEY ? "active" : "done"}
            summary={saved ? "Thanks" : "Skipped"}
          >
            {survey ? (
              <OnboardingStep
                survey={survey}
                baseUrl={surveyBaseUrl}
                hidden={{
                  instance_type: config.edition,
                  instance_domain: window.location.hostname,
                }}
                initialAnswers={
                  saved
                    ? mergeAnswers(emptyAnswers(survey), saved.answers)
                    : emptyAnswers(survey, { name, email })
                }
                responseId={saved?.responseId ?? null}
                onDone={(answers, responseId) => {
                  const next = { responseId, answers };
                  writeSavedSurvey(userId, next);
                  setSaved(next);
                  setStep(STEP_NAMESPACE);
                }}
                onSkip={() => setStep(STEP_NAMESPACE)}
              />
            ) : (
              <p className="text-xs text-text-muted">Loading survey...</p>
            )}
          </TrailStep>
        )}
        <TrailStep
          number={namespaceStep}
          title="Create a namespace"
          state={step === STEP_NAMESPACE ? "active" : "upcoming"}
        >
          <div className="space-y-3">
            <p className="text-xs text-text-secondary">
              A namespace holds your devices and the people who can reach them.
              Its name shows up in every SSH address, so keep it short.
            </p>
            <NamespaceCreateForm />
            {showSurvey && (
              <button
                type="button"
                onClick={() => setStep(STEP_SURVEY)}
                className="text-2xs font-medium text-text-muted hover:text-text-secondary transition-colors"
              >
                Back to the survey
              </button>
            )}
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

  return <CreateNamespaceTrail eyebrow={eyebrow} />;
}
