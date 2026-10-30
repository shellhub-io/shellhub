import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { getConfig } from "@/env";
import { useSubmitOnboardingSurvey } from "@/hooks/useOnboardingSurvey";
import {
  answerFields,
  emptyAnswers,
  responseData,
  surveyAnswersSchema,
  type KnownContact,
  type OnboardingSurvey,
  type SurveyAnswers,
} from "@/utils/onboardingSurvey";
import OnboardingSurveyFields from "./OnboardingSurveyFields";

interface OnboardingSurveyFormProps {
  known?: KnownContact;
  onDone: () => void;
  onSkip: () => void;
}

function SurveyForm({
  survey,
  known,
  onDone,
  onSkip,
}: OnboardingSurveyFormProps & { survey: OnboardingSurvey }) {
  const { control, handleSubmit, trigger } = useForm<SurveyAnswers>({
    resolver: zodResolver(surveyAnswersSchema(survey)),
    defaultValues: emptyAnswers(survey, known),
  });
  const submit = useSubmitOnboardingSurvey();
  const anonymous = useWatch({ control, name: "anonymous" });
  const questions = survey.questions.filter(
    (q) => !(anonymous && q.kind === "consent"),
  );
  const [step, setStep] = useState(0);
  const current = Math.min(step, questions.length - 1);
  const last = current === questions.length - 1;

  const next = async () => {
    if (await trigger(answerFields(questions[current]))) setStep(current + 1);
  };

  const send = (answers: SurveyAnswers) => {
    const hidden = {
      instance_type: getConfig().edition,
      instance_domain: window.location.hostname,
    };
    submit.mutate(
      { surveyId: survey.id, data: responseData(survey, answers, hidden) },
      { onSuccess: onDone },
    );
  };

  return (
    <form
      onSubmit={(e) => {
        if (last) void handleSubmit(send)(e);
        else {
          e.preventDefault();
          void next();
        }
      }}
      className="space-y-6"
      noValidate
    >
      <div className="space-y-2">
        <p className="text-xs text-text-secondary">
          A few questions about you and what you are connecting.
        </p>
        <p className="font-mono text-2xs text-text-muted">
          Question {current + 1} of {questions.length}
        </p>
        <div
          className="h-1 rounded-full bg-border overflow-hidden"
          aria-hidden="true"
        >
          <div
            className="h-full rounded-full bg-primary transition-all"
            style={{ width: `${((current + 1) / questions.length) * 100}%` }}
          />
        </div>
      </div>

      <OnboardingSurveyFields
        questions={[questions[current]]}
        control={control}
      />

      {submit.isError && (
        <Callout variant="error">
          Your answers could not be sent. Try again, or skip the survey.
        </Callout>
      )}

      <div className="flex items-center justify-end gap-3">
        <Button variant="secondary" onClick={onSkip}>
          Skip survey
        </Button>
        {current > 0 && (
          <Button variant="secondary" onClick={() => setStep(current - 1)}>
            Back
          </Button>
        )}
        <Button
          type="submit"
          loading={submit.isPending}
          onClick={(e) => {
            if (e.detail > 1) e.preventDefault();
          }}
        >
          {last ? "Continue" : "Next"}
        </Button>
      </div>
    </form>
  );
}

/**
 * The onboarding survey form: the questions one at a time, each checked before Next moves on, Back
 * to the one before, and on the last question a Continue button that sends the answers and calls
 * `onDone` once they are sent. A double click on Next counts as one, so it cannot step onto the
 * last question and send it in one go. Skip survey, on every question, calls `onSkip` without
 * sending anything. The consent question is left out while the user answers anonymously. It starts from blanks with the contact fields filled from `known`. A null survey is
 * still loading.
 */
export default function OnboardingSurveyForm({
  survey,
  ...props
}: OnboardingSurveyFormProps & { survey: OnboardingSurvey | null }) {
  if (!survey) {
    return <p className="text-xs text-text-muted">Loading survey...</p>;
  }
  return <SurveyForm survey={survey} {...props} />;
}
