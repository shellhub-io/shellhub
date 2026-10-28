/**
 * Returns from the namespace form to the onboarding survey that came before it.
 */
export default function BackToSurveyButton({
  onClick,
}: {
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="text-2xs font-medium text-text-muted hover:text-text-secondary transition-colors"
    >
      Back to the survey
    </button>
  );
}
