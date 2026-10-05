import { FormEvent } from "react";
import { useForm, type FieldPath } from "react-hook-form";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useSubmitOnboardingSurvey } from "@/hooks/useOnboardingSurvey";
import OnboardingSurveyFields from "./OnboardingSurveyFields";
import {
  responseData,
  validateAnswers,
  type OnboardingSurvey,
  type SurveyAnswers,
} from "./onboardingSurvey";

/**
 * The survey step of setup: the questions, and a Continue that sends the answers before moving
 * on. It starts from the answers given last time, so stepping back from the account step and
 * returning keeps them, and it updates the response already sent rather than sending a second.
 */
export default function OnboardingStep({
  survey,
  baseUrl,
  hidden,
  initialAnswers,
  responseId,
  onDone,
  onSkip,
}: {
  survey: OnboardingSurvey;
  baseUrl: string;
  hidden: Record<string, string>;
  initialAnswers: SurveyAnswers;
  responseId: string | null;
  onDone: (answers: SurveyAnswers, responseId: string) => void;
  onSkip: () => void;
}) {
  const { control, getValues, setError, clearErrors } = useForm<SurveyAnswers>({
    defaultValues: initialAnswers,
  });
  const submit = useSubmitOnboardingSurvey(baseUrl);

  const onSubmit = (e: FormEvent) => {
    e.preventDefault();
    clearErrors();

    const answers = getValues();
    const errors = Object.entries(validateAnswers(survey, answers));
    if (errors.length > 0) {
      for (const [path, message] of errors) {
        setError(path as FieldPath<SurveyAnswers>, { message });
      }
      return;
    }

    submit.mutate(
      {
        surveyId: survey.id,
        responseId,
        data: responseData(survey, answers, hidden),
      },
      { onSuccess: (id) => onDone(answers, id) },
    );
  };

  return (
    <form onSubmit={onSubmit} className="space-y-6" noValidate>
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
