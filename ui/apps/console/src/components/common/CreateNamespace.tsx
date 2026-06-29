import { useState, useEffect, FormEvent } from "react";
import {
  useCreateNamespace,
  useSwitchNamespace,
} from "@/hooks/useNamespaceMutations";
import { getNamespaces } from "@/client";
import { getConfig } from "@/env";
import { isSdkError } from "@/api/errors";
import {
  CommandLineIcon,
  SparklesIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import AmbientBackground from "./AmbientBackground";
import CopyButton from "@/components/common/CopyButton";
import NamespaceNameField from "@/components/common/fields/NamespaceNameField";
import {
  NAMESPACE_NAME_MIN_LENGTH,
  validateNamespaceName,
} from "@/utils/validation";
import {
  Button,
  GithubIcon,
  Spinner,
} from "@shellhub/design-system/primitives";

/* ─── Cloud/Enterprise form ─── */
function CloudForm() {
  const [name, setName] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createNs = useCreateNamespace();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const err = validateNamespaceName(name);
    if (err) {
      setValidationError(err);
      return;
    }
    setValidationError(null);
    setSubmitError(null);
    try {
      await createNs.mutateAsync(name);
    } catch (caught) {
      if (isSdkError(caught)) {
        if (caught.status === 409) {
          setSubmitError("A namespace with this name already exists.");
        } else if (caught.status === 403) {
          setSubmitError(
            "You have reached the namespace limit or do not have permission.",
          );
        } else if (caught.status === 400) {
          setSubmitError("The namespace name is invalid.");
        } else {
          setSubmitError("An unexpected error occurred. Please try again.");
        }
      } else {
        setSubmitError("An unexpected error occurred. Please try again.");
      }
    }
  };

  const displayError = validationError ?? submitError ?? null;

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="w-full">
      <div className="flex items-center gap-2">
        <div className="flex-1">
          <NamespaceNameField
            id="create-namespace-name"
            value={name}
            onChange={(v) => {
              setName(v);
              setValidationError(null);
              setSubmitError(null);
              createNs.reset();
            }}
            error={displayError}
          />
        </div>
        <Button
          type="submit"
          loading={createNs.isPending}
          disabled={
            createNs.isPending || name.length < NAMESPACE_NAME_MIN_LENGTH
          }
          className="shrink-0"
        >
          {createNs.isPending ? "Creating..." : "Create"}
        </Button>
      </div>
    </form>
  );
}

/* ─── Community CLI instructions ─── */
function CopyBlock({ command }: { command: string }) {
  return (
    <div className="relative bg-background border border-border rounded-lg p-3.5 pr-11 font-mono text-xs text-text-secondary leading-relaxed">
      <span className="text-primary/60">$ </span>
      {command}
      <CopyButton text={command} className="absolute top-2.5 right-2.5" />
    </div>
  );
}

function CommunityInstructions() {
  const switchNs = useSwitchNamespace();
  const [ready, setReady] = useState(false);
  const [tenantId, setTenantId] = useState<string | null>(null);

  const addCmd = "./bin/cli member add <username> <namespace> <role>";
  const createCmd = "./bin/cli namespace create <namespace> <owner-username>";

  // Poll for namespace assignment every 5 seconds (without updating the store)
  useEffect(() => {
    const check = async () => {
      try {
        const { data } = await getNamespaces({
          query: { page: 1, per_page: 1 },
          throwOnError: true,
        });
        if (data.length > 0) {
          setReady(true);
          setTenantId(data[0].tenant_id);
        }
      } catch {
        // ignore
      }
    };
    const interval = setInterval(() => void check(), 5000);
    return () => clearInterval(interval);
  }, []);

  const handleContinue = () => {
    if (tenantId) void switchNs.mutateAsync({ tenantId });
  };

  return (
    <div className="w-full space-y-5">
      <p className="text-sm text-text-secondary leading-relaxed">
        Ask the instance administrator to create a new namespace for you, or add
        you to an existing one.
      </p>

      {/* Option 1: Create new */}
      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
          Create a new namespace
        </p>
        <CopyBlock command={createCmd} />
      </div>

      {/* Option 2: Add to existing */}
      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
          Or add to an existing namespace
        </p>
        <CopyBlock command={addCmd} />
        <p className="mt-1.5 text-2xs text-text-muted">
          Roles: <span className="text-text-secondary">observer</span>,{" "}
          <span className="text-text-secondary">operator</span>,{" "}
          <span className="text-text-secondary">administrator</span>
        </p>
      </div>

      <Button fullWidth disabled={!ready} onClick={handleContinue}>
        {ready ? (
          "You're in! Go to dashboard"
        ) : (
          <>
            <Spinner size="md" tone="onPrimary" />
            Waiting for namespace access...
          </>
        )}
      </Button>

      {/* Upgrade tip */}
      <div className="flex items-start gap-2.5 bg-primary/5 border border-primary/10 rounded-lg p-3">
        <SparklesIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
        <p className="text-2xs text-text-secondary leading-relaxed">
          <span className="font-medium text-text-primary">Tip:</span>{" "}
          <a
            href="https://www.shellhub.io/pricing"
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:text-primary-400 transition-colors"
          >
            ShellHub Cloud and Enterprise
          </a>{" "}
          let you create and manage namespaces directly from the UI.
        </p>
      </div>
    </div>
  );
}

/* ─── Main component ─── */
export default function CreateNamespace() {
  const canCreate = getConfig().cloud || getConfig().enterprise;

  return (
    <div className="relative w-full min-h-0 flex-1 flex overflow-auto">
      <AmbientBackground />

      <div className="w-full max-w-5xl mx-auto px-8 py-12 flex flex-col">
        {/* Hero */}
        <div className="text-center mb-12 animate-fade-in">
          <div className="animate-float mb-6 inline-block">
            <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center shadow-lg shadow-primary/10">
              <CommandLineIcon
                className="w-10 h-10 text-primary"
                strokeWidth={1.2}
              />
            </div>
          </div>

          <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
            Welcome to ShellHub
          </p>
          <h1 className="text-3xl font-bold text-text-primary mb-3">
            Set up your namespace
          </h1>
          <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
            You need a namespace to continue. A namespace groups your devices,
            team members, sessions, and security rules — all in one place.
          </p>
        </div>

        {/* Form / CLI card */}
        <div
          className="w-full max-w-xl mx-auto bg-card/80 border border-border rounded-2xl p-8 backdrop-blur-sm animate-slide-up"
          style={{ animationDelay: "200ms" }}
        >
          {canCreate ? <CloudForm /> : <CommunityInstructions />}
        </div>

        {/* Footer links */}
        <div
          className="flex items-center justify-center gap-6 mt-10 animate-fade-in"
          style={{ animationDelay: "800ms" }}
        >
          <a
            href="https://docs.shellhub.io/self-hosted/administration"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <BookOpenIcon className="w-3.5 h-3.5" />
            Documentation
          </a>
          <span className="w-px h-3 bg-border" />
          <a
            href="https://github.com/shellhub-io/shellhub"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <GithubIcon className="w-3.5 h-3.5" />
            Community
          </a>
        </div>
      </div>
    </div>
  );
}
