import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { getConfig } from "@/env";
import { useSubmitOnboardingSurvey } from "@/hooks/useOnboardingSurvey";
import {
  emptyAnswers,
  responseData,
  surveyAnswersSchema,
  type KnownContact,
  type OnboardingSurvey,
  type SurveyAnswers,
} from "@/utils/onboardingSurvey";
import OnboardingSurveyFields from "./OnboardingSurveyFields";

interface StepProps {
  initialAnswers: SurveyAnswers | null;
  known?: KnownContact;
  responseId: string | null;
  onDone: (answers: SurveyAnswers, responseId: string) => void;
  onSkip: () => void;
}

function SurveyForm({
  survey,
  initialAnswers,
  known,
  responseId,
  onDone,
  onSkip,
}: StepProps & { survey: OnboardingSurvey }) {
  const { control, handleSubmit } = useForm<SurveyAnswers>({
    resolver: zodResolver(surveyAnswersSchema(survey)),
    defaultValues: initialAnswers ?? emptyAnswers(survey, known),
  });
  const submit = useSubmitOnboardingSurvey();

  const send = (answers: SurveyAnswers) => {
    const hidden = {
      instance_type: getConfig().edition,
      instance_domain: window.location.hostname,
    };
    submit.mutate(
      {
        surveyId: survey.id,
        responseId,
        data: responseData(survey, answers, hidden, {
          update: responseId !== null,
        }),
      },
      { onSuccess: (id) => onDone(answers, id) },
    );
  };

  return (
    <form
      onSubmit={(e) => void handleSubmit(send)(e)}
      className="space-y-6"
      noValidate
    >
      <p className="text-xs text-text-secondary">
        A few questions about you and what you are connecting.
      </p>

      <OnboardingSurveyFields survey={survey} control={control} />

      {submit.isError && (
        <Callout variant="error">
          Your answers could not be sent. Try again, or skip the survey.
        </Callout>
      )}

      <div className="flex items-center justify-end gap-3">
        {(import.meta.env.DEV || submit.isError) && (
          <Button variant="secondary" onClick={onSkip}>
            Skip survey
          </Button>
        )}
        <Button type="submit" loading={submit.isPending}>
          Continue
        </Button>
      </div>
    </form>
  );
}

/**
 * The survey step of a first-run trail: the questions, and a Continue that sends the answers
 * before moving on. A null survey is still loading. It starts from `initialAnswers` when the
 * caller kept the last ones, so stepping back from the next step keeps them, and otherwise from
 * blanks with the contact fields filled from `known`. Given a `responseId`, it updates that
 * response rather than sending a second.
 */
export default function OnboardingStep({
  survey,
  ...props
}: StepProps & { survey: OnboardingSurvey | null }) {
  if (!survey) {
    return <p className="text-xs text-text-muted">Loading survey...</p>;
  }
  return <SurveyForm survey={survey} {...props} />;
}
