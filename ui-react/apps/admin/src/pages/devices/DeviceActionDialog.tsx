import { useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useDevicesStore } from "../../stores/devicesStore";
import { Device } from "../../types/device";
import axios from "axios";

/* ─── Action Dialog ─── */
function DeviceActionDialog({
  device,
  action,
  onClose,
}: {
  device: Device;
  action: "accept" | "reject" | "remove";
  onClose: () => void;
}) {
  const store = useDevicesStore();
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const config = {
    accept: {
      title: "Accept Device",
      description: "Do you want to accept",
      confirm: "Accept",
      color: "bg-accent-green/90 hover:bg-accent-green",
    },
    reject: {
      title: "Reject Device",
      description: "Do you want to reject",
      confirm: "Reject",
      color: "bg-accent-yellow/90 hover:bg-accent-yellow text-background",
    },
    remove: {
      title: "Remove Device",
      description: "Do you want to remove",
      confirm: "Remove",
      color: "bg-accent-red/90 hover:bg-accent-red",
    },
  }[action];

  const handleConfirm = async () => {
    setProcessing(true);
    setError(null);
    try {
      await store[action](device.uid);
      onClose();
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        const status = err.response?.status;
        if (action === "accept" && status === 402) {
          setError(
            "Couldn't accept the device. Check your billing status and try again.",
          );
        } else if (action === "accept" && status === 403) {
          setError(
            "You reached the maximum amount of accepted devices in this namespace.",
          );
        } else if (action === "accept" && status === 409) {
          setError(
            "A device with that name already exists. Rename it and try again.",
          );
        } else {
          setError(`Failed to ${action} device.`);
        }
      } else {
        setError(`Failed to ${action} device.`);
      }
      setProcessing(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center">
      <div
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      />
      <div className="relative bg-surface border border-border rounded-2xl w-full max-w-sm mx-4 p-6 shadow-2xl animate-slide-up">
        <h2 className="text-base font-semibold text-text-primary mb-2">
          {config.title}
        </h2>
        <p className="text-sm text-text-muted mb-1">
          {config.description}{" "}
          <span className="font-medium text-text-primary">{device.name}</span>?
        </p>
        {action === "remove" && (
          <p className="text-xs text-text-muted/70 mb-4">
            This action cannot be undone.
          </p>
        )}
        {action !== "remove" && <div className="mb-4" />}
        {error && (
          <p className="text-xs font-mono text-accent-red mb-4 flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <button
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={processing}
            className={`px-5 py-2.5 ${config.color} text-white rounded-lg text-sm font-semibold disabled:opacity-dim transition-all`}
          >
            {processing ? "Processing..." : config.confirm}
          </button>
        </div>
      </div>
    </div>
  );
}

export default DeviceActionDialog;
