import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAdminStats } from "@/hooks/useAdminStats";

export default function DeviceLimitBanner() {
  const {
    data: license,
    isLoading: licenseLoading,
    isError: licenseError,
  } = useAdminLicense();
  const {
    stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();

  const cap = license?.features.devices;
  const registered = stats?.registered_devices;

  const over =
    typeof cap === "number" &&
    cap >= 0 &&
    typeof registered === "number" &&
    Number.isFinite(registered) &&
    registered >= cap;

  const approaching =
    typeof cap === "number" &&
    cap > 0 &&
    !over &&
    typeof registered === "number" &&
    Number.isFinite(registered) &&
    registered / cap >= 0.9;

  const visible =
    !licenseLoading &&
    !licenseError &&
    !statsLoading &&
    !statsError &&
    license != null &&
    (over || approaching);

  const isErrorSeverity = over;

  const message = over
    ? "Your licensed device limit has been reached — new devices can't connect until you remove devices or contact ShellHub sales to raise the limit."
    : "You're approaching your licensed device limit — contact ShellHub sales to raise it before new devices are blocked.";

  return (
    <div
      aria-hidden={!visible ? true : undefined}
      {...(!visible ? { inert: "" } : {})}
      className={`grid transition-[grid-template-rows] duration-300 ease-out ${
        visible ? "grid-rows-[1fr]" : "grid-rows-[0fr]"
      }`}
    >
      <div className="overflow-hidden">
        <div
          className={`${isErrorSeverity ? "bg-accent-red/[0.06] border-accent-red/10" : "bg-accent-yellow/[0.06] border-accent-yellow/10"} px-5 py-1.5 flex items-center gap-2 border-b`}
          role={isErrorSeverity ? "alert" : "status"}
          aria-live={isErrorSeverity ? "assertive" : "polite"}
        >
          <span
            className={`inline-flex rounded-full h-1.5 w-1.5 shrink-0 ${isErrorSeverity ? "bg-accent-red" : "bg-accent-yellow"}`}
          />
          <p
            className={`text-xs font-mono ${isErrorSeverity ? "text-accent-red" : "text-accent-yellow"}`}
          >
            {message}
          </p>
        </div>
      </div>
    </div>
  );
}
