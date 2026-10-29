import { useState } from "react";
import { useForm, useWatch, type FieldPath } from "react-hook-form";
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
  type SurveyQuestion,
} from "@/utils/onboardingSurvey";
import OnboardingSurveyFields from "./OnboardingSurveyFields";

interface StepProps {
  initialAnswers: SurveyAnswers | null;
  known?: KnownContact;
  responseId: string | null;
  onDone: (answers: SurveyAnswers, responseId: string) => void;
  onSkip: () => void;
  finishLabel?: string;
}

function answerFields(question: SurveyQuestion): FieldPath<SurveyAnswers>[] {
  switch (question.kind) {
    case "choice":
      return [`choices.${question.id}`, `other.${question.id}`];
    case "contact":
      return [`contact.${question.id}`];
    case "consent":
      return [`consent.${question.id}`];
  }
}

function SurveyForm({
  survey,
  initialAnswers,
  known,
  responseId,
  onDone,
  onSkip,
  finishLabel = "Continue",
}: StepProps & { survey: OnboardingSurvey }) {
  const { control, handleSubmit, trigger } = useForm<SurveyAnswers>({
    resolver: zodResolver(surveyAnswersSchema(survey)),
    defaultValues: initialAnswers ?? emptyAnswers(survey, known),
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
        survey={{ ...survey, questions: [questions[current]] }}
        control={control}
      />

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
          {last ? finishLabel : "Next"}
        </Button>
      </div>
    </form>
  );
}

/**
 * The survey step of a first-run trail: the questions one at a time, each checked before Next moves
 * on, Back to the one before, and on the last question a button labelled `finishLabel` ("Continue"
 * by default) that sends the answers before moving on. A double click on that button counts as one,
 * so it cannot step onto the last question and send it in one go. The consent question is left out
 * while the user answers anonymously. A null survey is still loading. It starts from
 * `initialAnswers` when the caller kept the last ones, so stepping back from the next step keeps
 * them, and otherwise from blanks with the contact fields filled from `known`. Given a
 * `responseId`, it updates that response rather than sending a second.
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
