import { useMemo } from "react";
import { getConfig } from "@/env";
import { useAdminLicense } from "@/hooks/useAdminLicense";

export default function LicenseBanner() {
  const { data, isLoading, isError, dataUpdatedAt } = useAdminLicense();
  const isEnterprise = getConfig().enterprise && !getConfig().cloud;

  // data === undefined: query not enabled (non-admin) — banner stays hidden
  // data === null:      no license installed (400 normalized by hook)
  // data object:        license found, check flags
  const noLicense = data === null;
  const expired = data != null && data.expired && !data.grace_period;
  const gracePeriod = data != null && data.expired && data.grace_period;
  const aboutToExpire = data != null && !data.expired && data.about_to_expire;

  const daysUntilExpiration = useMemo(() => {
    if (data == null || data.expires_at <= 0) return null;
    return Math.ceil((data.expires_at - dataUpdatedAt / 1000) / 86400);
  }, [data, dataUpdatedAt]);

  const visible = isEnterprise && !isLoading && !isError
    && (noLicense || expired || gracePeriod || aboutToExpire);

  const isErrorSeverity = noLicense;

  const message = (() => {
    if (noLicense) return "No license uploaded. This instance requires a valid license to operate properly.";
    if (expired) return "Your license has expired. The instance will not function properly until a new license is uploaded.";
    if (gracePeriod) return "Your license has expired and is in the grace period. Upload a new license before it stops working.";
    if (aboutToExpire) {
      return daysUntilExpiration !== null
        ? `Your license expires in ${daysUntilExpiration} day${daysUntilExpiration === 1 ? "" : "s"}. Upload a new license soon to avoid interruption.`
        : "Your license is about to expire. Upload a new license soon to avoid interruption.";
    }
    return "";
  })();

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
          className={`${
            isErrorSeverity
              ? "bg-accent-red/[0.06] border-b border-accent-red/10"
              : "bg-accent-yellow/[0.06] border-b border-accent-yellow/10"
          } px-5 py-1.5 flex items-center justify-between gap-2`}
          role={isErrorSeverity ? "alert" : "status"}
          aria-live={isErrorSeverity ? "assertive" : "polite"}
        >
          <div className="flex items-center gap-2 min-w-0">
            <span
              className={`inline-flex rounded-full h-1.5 w-1.5 shrink-0 ${
                isErrorSeverity ? "bg-accent-red" : "bg-accent-yellow"
              }`}
            />
            <p
              className={`text-2xs font-mono ${
                isErrorSeverity ? "text-accent-red" : "text-accent-yellow"
              }`}
            >
              {message}
            </p>
          </div>
          <a
            href="/admin/license"
            className={`text-2xs font-mono font-semibold underline shrink-0 ${
              isErrorSeverity ? "text-accent-red" : "text-accent-yellow"
            }`}
          >
            Upload license
          </a>
        </div>
      </div>
    </div>
  );
}
