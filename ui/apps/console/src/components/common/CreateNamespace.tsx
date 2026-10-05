import { useNamespaceArrival } from "@/hooks/useNamespaceArrival";
import { useNamespaceCreateForm } from "@/hooks/useNamespaceCreateForm";
import { SparklesIcon } from "@heroicons/react/24/outline";
import CopyButton from "@/components/common/CopyButton";
import NamespaceNameField from "@/components/common/fields/NamespaceNameField";
import { NAMESPACE_NAME_MIN_LENGTH } from "@/utils/validation";
import { Button, Spinner } from "@shellhub/design-system/primitives";

/**
 * The namespace name form. Validates as the user types against the same rules the server holds,
 * so the requirements are visible before the request rather than after it.
 */
export function NamespaceCreateForm() {
  const form = useNamespaceCreateForm();

  return (
    <form onSubmit={(e) => void form.submit(e)} className="w-full">
      <div className="flex items-center gap-2">
        <div className="flex-1">
          <NamespaceNameField
            id="create-namespace-name"
            value={form.name}
            onChange={form.changeName}
            error={form.error}
          />
        </div>
        <Button
          type="submit"
          loading={form.isPending}
          disabled={
            form.isPending || form.name.length < NAMESPACE_NAME_MIN_LENGTH
          }
          className="shrink-0"
        >
          {form.isPending ? "Creating..." : "Create"}
        </Button>
      </div>
    </form>
  );
}

function CopyBlock({ command }: { command: string }) {
  return (
    <div className="relative bg-background border border-border rounded-lg p-3.5 pr-11 font-mono text-xs text-text-secondary leading-relaxed">
      <span className="text-primary/60">$ </span>
      {command}
      <CopyButton text={command} className="absolute top-2.5 right-2.5" />
    </div>
  );
}

/**
 * What community edition shows instead of the form: creating a namespace there is a server-side
 * step, so this explains it and then offers to switch into the namespace once it exists, or
 * switches in by itself with `autoEnter`.
 */
export function CommunityInstructions({
  autoEnter = false,
}: {
  autoEnter?: boolean;
}) {
  const { ready, enter, pending, failed } = useNamespaceArrival({ autoEnter });

  const addCmd = "./bin/cli member add <username> <namespace> <role>";

  return (
    <div className="w-full space-y-5">
      <p className="text-sm text-text-secondary leading-relaxed">
        Ask the instance administrator to add you to the namespace.
      </p>

      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2">
          Add a member to the namespace
        </p>
        <CopyBlock command={addCmd} />
        <p className="mt-1.5 text-2xs text-text-muted">
          Roles: <span className="text-text-secondary">observer</span>,{" "}
          <span className="text-text-secondary">operator</span>,{" "}
          <span className="text-text-secondary">administrator</span>
        </p>
      </div>

      {failed && (
        <p role="alert" className="text-xs text-accent-red">
          You were added, but getting into the namespace failed.
        </p>
      )}

      <Button fullWidth disabled={!ready || pending} onClick={enter}>
        {failed ? (
          "Try again"
        ) : ready ? (
          "You're in! Go to dashboard"
        ) : (
          <>
            <Spinner size="md" tone="onPrimary" />
            Waiting for namespace access...
          </>
        )}
      </Button>

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
