import { useId } from "react";
import { useWatch, type Control } from "react-hook-form";
import {
  FormCheckboxField,
  FormInputField,
  FormRadioGroupField,
} from "@/components/common/fields/rhf";
import RadioPill from "@/components/common/fields/RadioPill";
import {
  isOtherChoice,
  type OnboardingSurvey,
  type SurveyAnswers,
  type SurveyQuestion,
} from "./onboardingSurvey";

function QuestionHeading({
  id,
  question,
}: {
  id: string;
  question: SurveyQuestion;
}) {
  return (
    <div className="mb-2">
      <p id={id} className="text-sm font-medium text-text-primary">
        {question.headline}
        {question.required && (
          <span className="text-text-muted" aria-hidden="true">
            {" "}
            *
          </span>
        )}
      </p>
      {question.subheader && (
        <p className="text-xs text-text-muted mt-0.5">{question.subheader}</p>
      )}
    </div>
  );
}

function ChoiceQuestion({
  question,
  control,
}: {
  question: Extract<SurveyQuestion, { kind: "choice" }>;
  control: Control<SurveyAnswers>;
}) {
  const headingId = useId();
  const selected = useWatch({ control, name: `choices.${question.id}` });
  const other = question.choices.find(isOtherChoice);

  return (
    <fieldset>
      <QuestionHeading id={headingId} question={question} />
      <FormRadioGroupField<SurveyAnswers, string>
        name={`choices.${question.id}`}
        control={control}
        labelledBy={headingId}
        containerClassName="flex flex-wrap gap-2"
      >
        {question.choices.map((choice) => (
          <RadioPill key={choice.id} value={choice.id} label={choice.label} />
        ))}
      </FormRadioGroupField>
      {other && selected === other.id && (
        <div className="mt-2">
          <FormInputField<SurveyAnswers>
            id={`${question.id}-other`}
            label={`${other.label} (please specify)`}
            hideLabel
            name={`other.${question.id}`}
            control={control}
            placeholder="Please specify"
            maxLength={200}
          />
        </div>
      )}
    </fieldset>
  );
}

function ContactQuestion({
  question,
  control,
}: {
  question: Extract<SurveyQuestion, { kind: "contact" }>;
  control: Control<SurveyAnswers>;
}) {
  const headingId = useId();

  return (
    <fieldset aria-labelledby={headingId}>
      <QuestionHeading id={headingId} question={question} />
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {question.fields.map((field) => (
          <FormInputField<SurveyAnswers>
            key={field.name}
            id={`${question.id}-${field.name}`}
            label={field.placeholder || field.name}
            name={`contact.${question.id}.${field.name}`}
            control={control}
            type={field.name === "email" ? "email" : "text"}
            maxLength={200}
          />
        ))}
      </div>
    </fieldset>
  );
}

function ConsentQuestion({
  question,
  control,
}: {
  question: Extract<SurveyQuestion, { kind: "consent" }>;
  control: Control<SurveyAnswers>;
}) {
  const headingId = useId();

  return (
    <fieldset aria-labelledby={headingId}>
      <QuestionHeading id={headingId} question={question} />
      <FormCheckboxField<SurveyAnswers>
        id={question.id}
        name={`consent.${question.id}`}
        control={control}
        label={question.label}
        labelSize="sm"
      />
    </fieldset>
  );
}

/**
 * The onboarding survey's questions as console form fields, bound to a form the caller owns so the
 * answers outlive this component when the user steps away and back.
 */
export default function OnboardingSurveyFields({
  survey,
  control,
}: {
  survey: OnboardingSurvey;
  control: Control<SurveyAnswers>;
}) {
  return (
    <div className="space-y-6">
      {survey.questions.map((question) => {
        switch (question.kind) {
          case "choice":
            return (
              <ChoiceQuestion
                key={question.id}
                question={question}
                control={control}
              />
            );
          case "contact":
            return (
              <ContactQuestion
                key={question.id}
                question={question}
                control={control}
              />
            );
          case "consent":
            return (
              <ConsentQuestion
                key={question.id}
                question={question}
                control={control}
              />
            );
        }
      })}
    </div>
  );
}
