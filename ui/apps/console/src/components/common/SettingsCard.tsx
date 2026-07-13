import React from "react";
import { cn } from "@shellhub/design-system/cn";

type SettingsCardProps = {
  title: string;
  children: React.ReactNode;
  danger?: boolean;
};

export default function SettingsCard({ title, children, danger }: SettingsCardProps) {
  return (
    <div
      className={cn("bg-card border rounded-xl overflow-hidden", danger ? "border-accent-red/20 border-l-2 border-l-accent-red/40" : "border-border")}
    >
      <div
        className={cn("px-5 py-3.5 border-b", danger ? "border-accent-red/10" : "border-border")}
      >
        <h3
          className={cn("text-sm font-semibold", danger ? "text-accent-red" : "text-text-primary")}
        >
          {title}
        </h3>
      </div>
      <div className="divide-y divide-border">{children}</div>
    </div>
  );
}
