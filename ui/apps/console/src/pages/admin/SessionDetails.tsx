import { useParams, Link } from "react-router-dom";
import {
  CommandLineIcon,
  ExclamationCircleIcon,
  CheckCircleIcon,
  MinusCircleIcon,
} from "@heroicons/react/24/outline";
import { useAdminSessionDetail } from "@/hooks/useAdminSessionDetail";
import Breadcrumb from "@/components/common/Breadcrumb";
import { formatDateFull } from "@/utils/date";
import { sessionType } from "@/utils/session";
import PageLoader from "@/components/common/PageLoader";
import { Card } from "@shellhub/design-system/primitives";

function Field({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="py-3 border-b border-border/50 last:border-0">
      <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-1">
        {label}
      </p>
      <div className="text-sm text-text-primary">{children}</div>
    </div>
  );
}

function BoolField({
  value,
  falseColor = "text-accent-red",
}: {
  value: boolean;
  falseColor?: string;
}) {
  return (
    <span
      className={`flex items-center gap-1.5 text-sm ${value ? "text-accent-green" : falseColor}`}
    >
      {value ? (
        <CheckCircleIcon className="w-4 h-4" strokeWidth={2} />
      ) : (
        <MinusCircleIcon className="w-4 h-4" strokeWidth={2} />
      )}
      {value ? "Yes" : "No"}
    </span>
  );
}

export default function AdminSessionDetails() {
  const { uid = "" } = useParams<{ uid: string }>();
  const { session, isLoading, error } = useAdminSessionDetail(uid);

  if (isLoading) {
    return <PageLoader label="Loading session" padding="fill" />;
  }

  if (error || !session) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center" role="alert">
          <ExclamationCircleIcon className="w-10 h-10 text-accent-red mx-auto mb-3" />
          <p className="text-sm font-medium text-text-primary">
            Session not found
          </p>
          <p className="text-2xs text-text-muted mt-1">
            {error?.message ??
              "The session may have been removed or the ID is invalid."}
          </p>
        </div>
      </div>
    );
  }

  const type = sessionType(session);

  return (
    <div>
      <Breadcrumb
        items={[
          { label: "Sessions", to: "/admin/sessions" },
          { label: session.uid.slice(0, 8), title: session.uid },
        ]}
      />

      <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
            <CommandLineIcon className="w-7 h-7" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary leading-tight">
              Session Details
            </h1>
            <p className="text-sm text-text-muted mt-1 max-w-xl">
              Detailed information about the selected session.
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <span
            className={`w-2 h-2 rounded-full inline-block shrink-0 ${
              session.active
                ? "bg-accent-green shadow-[0_0_6px_rgba(130,165,104,0.4)]"
                : "bg-text-muted/40"
            }`}
            aria-label={session.active ? "Active" : "Inactive"}
          />
          <code className="text-xs font-mono text-text-muted break-all">
            {session.uid}
          </code>
        </div>
      </div>

      <Card className="rounded-lg overflow-hidden animate-fade-in">
        <div className="grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border">
          <div className="px-6 py-2">
            <Field label="UID">
              <code className="text-xs font-mono text-text-secondary break-all">
                {session.uid}
              </code>
            </Field>

            {session.device && (
              <Field label="Device">
                <Link
                  to={`/admin/devices/${session.device.uid}`}
                  className="text-primary hover:underline text-sm"
                >
                  {session.device.name || session.device.uid}
                </Link>
              </Field>
            )}

            <Field label="Username">
              <code className="text-xs font-mono">{session.username}</code>
            </Field>

            <Field label="IP Address">
              <code className="text-xs font-mono text-text-muted bg-surface px-1.5 py-0.5 rounded">
                {session.ip_address}
              </code>
            </Field>

            <Field label="Type">
              {type ? (
                <span
                  className={`inline-flex items-center px-2 py-0.5 text-2xs font-mono font-semibold rounded border ${type.color}`}
                >
                  {type.label}
                </span>
              ) : (
                <span className="text-text-secondary capitalize">
                  {session.type}
                </span>
              )}
            </Field>

            <Field label="Terminal">
              <span className="text-text-secondary">
                {session.term === "none" || !session.term ? "—" : session.term}
              </span>
            </Field>
          </div>

          <div className="px-6 py-2">
            {session.device?.namespace && (
              <Field label="Namespace">
                <span className="text-text-secondary">
                  {session.device.namespace}
                </span>
              </Field>
            )}

            <Field label="Authenticated">
              <BoolField
                value={session.authenticated}
                falseColor="text-accent-red"
              />
            </Field>

            <Field label="Recorded">
              <BoolField
                value={session.recorded}
                falseColor="text-text-secondary"
              />
            </Field>

            <Field label="Started At">
              <span className="text-text-secondary">
                {formatDateFull(session.started_at)}
              </span>
            </Field>

            <Field label="Last Seen">
              <span className="text-text-secondary">
                {formatDateFull(session.last_seen)}
              </span>
            </Field>
          </div>
        </div>
      </Card>
    </div>
  );
}
