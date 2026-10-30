import { Outlet } from "react-router-dom";
import { isEnterprise } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import {
  useOnboardingSurvey,
  useRecordOnboardingSurvey,
} from "@/hooks/useOnboardingSurvey";
import FirstRunLayout from "./FirstRunLayout";
import OnboardingSurveyForm from "./survey/OnboardingSurveyForm";

/**
 * Asks the onboarding survey once, before any signed-in page, of a user the server still holds as
 * pending, on community and cloud. Answering or skipping records the outcome and lets the user on
 * to the page they asked for. Every other user passes straight through, and so does a pending one
 * when the survey cannot be shown, with nothing recorded, so a later sign-in asks again. An
 * administrator signed in as the user passes through too, since the outcome is the user's.
 */
export default function OnboardingSurveyGate() {
  const pending = useAuthStore(
    (s) => s.onboardingSurvey === "pending" && !s.impersonated,
  );
  const name = useAuthStore((s) => s.name);
  const email = useAuthStore((s) => s.email);
  const { visible, survey } = useOnboardingSurvey(pending && !isEnterprise());
  const record = useRecordOnboardingSurvey();

  if (!visible) return <Outlet />;

  return (
    <FirstRunLayout
      eyebrow={`Welcome, ${name || email || "back"}`}
      title="Tell us about you"
      lead="Before you start, a few questions. You only see them once."
      signedIn
      inConsole={false}
    >
      <OnboardingSurveyForm
        survey={survey}
        known={{ name, email }}
        onDone={() => record("answered")}
        onSkip={() => record("skipped")}
      />
    </FirstRunLayout>
  );
}
