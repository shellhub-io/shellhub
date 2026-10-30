import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { isSdkError } from "@/api/errors";
import { consumePendingDeviceCode } from "@/utils/navigation";
import { validateNamespaceName } from "@/utils/validation";
import { useCreateNamespace } from "@/hooks/useNamespaceRequests";

const GENERIC_ERROR = "An unexpected error occurred. Please try again.";

const SUBMIT_ERRORS: Record<number, string> = {
  400: "The namespace name is invalid.",
  403: "You have reached the namespace limit or do not have permission.",
  409: "A namespace with this name already exists.",
};

/**
 * The text for a rejected namespace creation: one message each for a 400, a 403 and a 409, and a
 * generic one for any other status or for an error that is not an API response.
 */
export function namespaceCreateError(caught: unknown): string {
  return (isSdkError(caught) && SUBMIT_ERRORS[caught.status]) || GENERIC_ERROR;
}

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
  const navigate = useNavigate();

  const landInNewNamespace = () => {
    const pendingCode = consumePendingDeviceCode();
    void navigate(
      pendingCode
        ? `/accept-device?code=${encodeURIComponent(pendingCode)}`
        : "/dashboard",
    );
  };

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
      await createNs.mutateAsync({ name, land: landInNewNamespace });
      onCreated?.();
    } catch (caught) {
      setSubmitError(namespaceCreateError(caught));
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
