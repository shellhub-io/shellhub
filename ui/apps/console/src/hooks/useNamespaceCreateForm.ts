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

function namespaceCreateError(caught: unknown): string {
  return (isSdkError(caught) && SUBMIT_ERRORS[caught.status]) || GENERIC_ERROR;
}

/**
 * Drives a namespace name form: it starts from `initialName`, validates before the request, keeps
 * the field and its error in step, and shows whatever `create` rejects with on the name field: one
 * message each for a 400, a 403 and a 409, and a generic one for any other status or for an error
 * that is not an API response. That includes a namespace created but not entered because its token
 * could not be issued: the message is the one for the token request's status, usually the generic
 * one, while the namespace exists, so submitting the same name again gets the 409. A failure in a
 * step `create` runs after that, such as accepting a device, belongs elsewhere on the page, so
 * `create` reports it itself rather than rejecting. Without `create`, the form creates the
 * namespace, enters it and lands on the device code waiting to be accepted, or on the dashboard.
 * `isPending` holds while `create` runs, and `onCreated` runs once it resolves.
 */
export function useNamespaceCreateForm({
  initialName = "",
  create,
  onCreated,
}: {
  initialName?: string;
  create?: (name: string) => Promise<unknown>;
  onCreated?: () => void;
} = {}) {
  const [name, setName] = useState(initialName);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);
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

  const createAndLand = (newName: string) =>
    createNs.mutateAsync({ name: newName, land: landInNewNamespace });

  const changeName = (value: string) => {
    setName(value);
    setValidationError(null);
    setSubmitError(null);
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
    setIsPending(true);
    try {
      await (create ?? createAndLand)(name);
      onCreated?.();
    } catch (caught) {
      setSubmitError(namespaceCreateError(caught));
    } finally {
      setIsPending(false);
    }
  };

  return {
    name,
    changeName,
    submit,
    error: validationError ?? submitError,
    isPending,
  };
}
