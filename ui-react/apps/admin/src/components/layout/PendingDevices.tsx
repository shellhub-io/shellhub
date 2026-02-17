import { useState, useRef, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import {
  XMarkIcon,
  CheckIcon,
  CheckCircleIcon,
  CpuChipIcon,
  ArrowRightIcon,
} from "@heroicons/react/24/outline";
import { getDevices, acceptDevice, rejectDevice } from "../../api/devices";
import { Device } from "../../types/device";
import { useClickOutside } from "../../hooks/useClickOutside";

export default function PendingDevices() {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [devices, setDevices] = useState<Device[]>([]);
  const [count, setCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [acting, setActing] = useState<string | null>(null);
  const [flash, setFlash] = useState<{
    uid: string;
    action: "accepted" | "rejected";
  } | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const fetchPending = useCallback(async () => {
    try {
      setLoading(true);
      const { data: d, totalCount } = await getDevices(1, 5, "pending");
      setDevices(d);
      setCount(totalCount);
    } catch {
      /* silent */
    } finally {
      setLoading(false);
    }
  }, []);

  // Poll for pending count every 30s
  useEffect(() => {
    fetchPending();
    const id = setInterval(fetchPending, 30000);
    return () => clearInterval(id);
  }, [fetchPending]);

  // Refetch when opened
  useEffect(() => {
    if (open) fetchPending();
  }, [open, fetchPending]);

  useClickOutside(containerRef, () => setOpen(false));

  const handleAction = async (uid: string, action: "accepted" | "rejected") => {
    setActing(uid);
    try {
      if (action === "accepted") await acceptDevice(uid);
      else await rejectDevice(uid);

      setFlash({ uid, action });
      setTimeout(async () => {
        setFlash(null);
        await fetchPending();
        setActing(null);
      }, 600);
    } catch {
      setActing(null);
    }
  };

  return (
    <div ref={containerRef} className="relative">
      {/* Trigger */}
      <button
        onClick={() => setOpen(!open)}
        className="relative flex items-center justify-center w-8 h-8 rounded-lg border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
      >
        <CpuChipIcon className="w-[18px] h-[18px] text-text-secondary" />

        {/* Badge */}
        {count > 0 && (
          <span className="absolute -top-0.5 -right-0.5 min-w-[16px] h-4 px-1 flex items-center justify-center rounded-full bg-accent-yellow text-3xs font-bold font-mono text-background leading-none">
            {count > 99 ? "99+" : count}
          </span>
        )}
      </button>

      {/* Dropdown */}
      {open && (
        <div className="absolute top-full right-0 mt-1.5 w-80 bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down">
          {/* Header */}
          <div className="flex items-center justify-between px-4 py-3 border-b border-border">
            <div className="flex items-center gap-2">
              <h3 className="text-xs font-semibold text-text-primary">
                Pending Devices
              </h3>
              {count > 0 && (
                <span className="text-2xs font-mono font-semibold text-accent-yellow bg-accent-yellow/10 border border-accent-yellow/20 px-1.5 py-0.5 rounded">
                  {count}
                </span>
              )}
            </div>
            <button
              onClick={() => setOpen(false)}
              className="p-1 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
            >
              <XMarkIcon className="w-3.5 h-3.5" strokeWidth={2} />
            </button>
          </div>

          {/* Body */}
          <div className="max-h-[280px] overflow-y-auto">
            {loading && devices.length === 0 ? (
              <div className="flex items-center justify-center py-10">
                <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
              </div>
            ) : devices.length === 0 ? (
              <div className="py-10 text-center">
                <CheckCircleIcon
                  className="w-8 h-8 text-text-muted/20 mx-auto mb-2"
                  strokeWidth={1}
                />
                <p className="text-xs text-text-muted">No pending devices</p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {devices.map((d) => {
                  const isActing = acting === d.uid;
                  const isFlashed = flash?.uid === d.uid;

                  return (
                    <div
                      key={d.uid}
                      className={`px-4 py-3 transition-all duration-300 ${
                        isFlashed
                          ? flash.action === "accepted"
                            ? "bg-accent-green/10"
                            : "bg-accent-red/10"
                          : "hover:bg-hover-subtle"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0 flex-1">
                          <p className="text-sm font-medium text-text-primary truncate">
                            {d.name}
                          </p>
                          <div className="flex items-center gap-2 mt-1">
                            <span className="text-2xs font-mono text-text-muted truncate">
                              {d.identity?.mac || d.info?.id || "\u2014"}
                            </span>
                            {d.info?.pretty_name && (
                              <>
                                <span className="w-px h-2.5 bg-border shrink-0" />
                                <span className="text-2xs text-text-muted truncate">
                                  {d.info.pretty_name}
                                </span>
                              </>
                            )}
                          </div>
                        </div>

                        {/* Actions */}
                        {isFlashed ? (
                          <span
                            className={`text-2xs font-mono font-semibold ${
                              flash.action === "accepted"
                                ? "text-accent-green"
                                : "text-accent-red"
                            }`}
                          >
                            {flash.action === "accepted"
                              ? "Accepted"
                              : "Rejected"}
                          </span>
                        ) : (
                          <div className="flex items-center gap-1 shrink-0">
                            <button
                              onClick={() => handleAction(d.uid, "accepted")}
                              disabled={isActing}
                              className="p-1.5 rounded-md text-text-muted hover:text-accent-green hover:bg-accent-green/10 transition-colors disabled:opacity-soft"
                              title="Accept"
                            >
                              {isActing ? (
                                <span className="block w-3.5 h-3.5 border-2 border-text-muted/30 border-t-text-muted rounded-full animate-spin" />
                              ) : (
                                <CheckIcon
                                  className="w-3.5 h-3.5"
                                  strokeWidth={2.5}
                                />
                              )}
                            </button>
                            <button
                              onClick={() => handleAction(d.uid, "rejected")}
                              disabled={isActing}
                              className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-colors disabled:opacity-soft"
                              title="Reject"
                            >
                              <XMarkIcon
                                className="w-3.5 h-3.5"
                                strokeWidth={2.5}
                              />
                            </button>
                          </div>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Footer */}
          {count > 0 && (
            <div className="px-4 py-2.5 border-t border-border">
              <button
                onClick={() => {
                  setOpen(false);
                  navigate("/devices");
                }}
                className="w-full flex items-center justify-center gap-1.5 text-xs font-medium text-primary hover:text-primary/80 transition-colors"
              >
                View all pending devices
                <ArrowRightIcon className="w-3 h-3" strokeWidth={2.5} />
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
