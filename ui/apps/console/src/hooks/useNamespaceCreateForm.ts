import { useState, type FormEvent } from "react";
import { isSdkError } from "@/api/errors";
import { validateNamespaceName } from "@/utils/validation";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";

const GENERIC_ERROR = "An unexpected error occurred. Please try again.";

const SUBMIT_ERRORS: Record<number, string> = {
  400: "The namespace name is invalid.",
  403: "You have reached the namespace limit or do not have permission.",
  409: "A namespace with this name already exists.",
};

/**
 * Drives the namespace name form shared by the create-namespace screen and dialog: it validates
 * before the request, keeps the field and its error in step, and turns a rejected creation into
 * text keyed by status. `onCreated` runs only after a successful creation.
 */
export function useNamespaceCreateForm(onCreated?: () => void) {
  const [name, setName] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createNs = useCreateNamespace();

  const changeName = (value: string) => {
    setName(value);
    setValidationError(null);
    setSubmitError(null);
    createNs.reset();
  };

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    const invalid = validateNamespaceName(name);
    if (invalid) {
      setValidationError(invalid);
      return;
    }
    setValidationError(null);
    setSubmitError(null);
    try {
      await createNs.mutateAsync(name);
      onCreated?.();
    } catch (caught) {
      setSubmitError(
        (isSdkError(caught) && SUBMIT_ERRORS[caught.status]) || GENERIC_ERROR,
      );
    }
  };

  return {
    name,
    changeName,
    submit,
    error: validationError ?? submitError,
    isPending: createNs.isPending,
  };
}
