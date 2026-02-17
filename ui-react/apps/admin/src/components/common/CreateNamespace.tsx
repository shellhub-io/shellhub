import { useState, useEffect, FormEvent } from "react";
import { useNamespacesStore } from "../../stores/namespacesStore";
import { getNamespaces } from "../../api/namespaces";
import { getConfig } from "../../env";
import {
  ExclamationCircleIcon,
  CheckIcon,
  CommandLineIcon,
  ClipboardDocumentIcon,
  SparklesIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import AmbientBackground from "./AmbientBackground";

const NAME_REGEX = /^[a-z0-9][a-z0-9-]*[a-z0-9]$/;

function validate(name: string): string | null {
  if (name.length < 3) return "Name must be at least 3 characters";
  if (name.length > 30) return "Name must be at most 30 characters";
  if (!NAME_REGEX.test(name))
    return "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)";
  return null;
}

/* ─── Cloud/Enterprise form ─── */
function CloudForm() {
  const [name, setName] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);
  const { createNamespace, loading, error } = useNamespacesStore();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const err = validate(name);
    if (err) {
      setValidationError(err);
      return;
    }
    setValidationError(null);
    try {
      await createNamespace(name);
    } catch {
      // error is set in store
    }
  };

  const displayError = validationError || error;

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <label className="block text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
        Namespace Name
      </label>

      <div className="flex gap-2">
        <input
          type="text"
          value={name}
          onChange={(e) => {
            setName(e.target.value.toLowerCase());
            setValidationError(null);
          }}
          placeholder="my-namespace"
          maxLength={30}
          autoFocus
          className="flex-1 px-4 py-3 bg-background border border-border rounded-lg text-sm text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all duration-200"
        />
        <button
          type="submit"
          disabled={loading || name.length < 3}
          className="px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all duration-200 shrink-0"
        >
          {loading ? (
            <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" />
          ) : (
            "Create"
          )}
        </button>
      </div>

      {displayError && (
        <p className="mt-2.5 text-xs font-mono text-accent-red animate-slide-down flex items-center gap-1.5">
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {displayError}
        </p>
      )}

      <p className="mt-2.5 text-2xs text-text-muted">
        3–30 characters · lowercase letters, numbers, and hyphens only
      </p>
    </form>
  );
}

/* ─── Community CLI instructions ─── */
function CopyBlock({ command }: { command: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(command);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative bg-background border border-border rounded-lg p-3.5 pr-11 font-mono text-xs text-text-secondary leading-relaxed">
      <span className="text-primary/60">$ </span>
      {command}
      <button
        onClick={handleCopy}
        className="absolute top-2.5 right-2.5 p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
        title="Copy command"
      >
        {copied ? (
          <CheckIcon
            className="w-3.5 h-3.5 text-accent-green"
            strokeWidth={2}
          />
        ) : (
          <ClipboardDocumentIcon className="w-3.5 h-3.5" strokeWidth={2} />
        )}
      </button>
    </div>
  );
}

function CommunityInstructions() {
  const { switchNamespace } = useNamespacesStore();
  const [ready, setReady] = useState(false);
  const [tenantId, setTenantId] = useState<string | null>(null);

  const addCmd = "./bin/cli member add <username> <namespace> <role>";
  const createCmd = "./bin/cli namespace create <namespace> <owner-username>";

  // Poll for namespace assignment every 5 seconds (without updating the store)
  useEffect(() => {
    const check = async () => {
      try {
        const ns = await getNamespaces(1, 1);
        if (ns.length > 0) {
          setReady(true);
          setTenantId(ns[0].tenant_id);
        }
      } catch {
        // ignore
      }
    };
    const interval = setInterval(check, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleContinue = () => {
    if (tenantId) {
      switchNamespace(tenantId);
    }
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

      <button
        disabled={!ready}
        onClick={handleContinue}
        className={`w-full flex items-center justify-center gap-2.5 px-5 py-3 rounded-lg text-sm font-semibold transition-all duration-200 ${
          ready
            ? "bg-primary hover:bg-primary-600 text-white cursor-pointer"
            : "bg-primary/30 text-white/50 cursor-not-allowed"
        }`}
      >
        {ready ? (
          "You're in! Go to dashboard"
        ) : (
          <>
            <span className="w-4 h-4 border-2 border-white/20 border-t-white/50 rounded-full animate-spin" />
            Waiting for namespace access...
          </>
        )}
      </button>

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

      <div className="relative z-10 w-full max-w-5xl mx-auto px-8 py-12 flex flex-col">
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
            <svg
              className="w-3.5 h-3.5"
              fill="currentColor"
              viewBox="0 0 24 24"
            >
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            Community
          </a>
        </div>
      </div>
    </div>
  );
}
