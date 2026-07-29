import React from "react";

type SettingsRowProps = {
  icon: React.ReactNode;
  title: string;
  description: React.ReactNode;
  badge?: React.ReactNode;
  children: React.ReactNode;
};

export default function SettingsRow({
  icon,
  title,
  description,
  badge,
  children,
}: SettingsRowProps) {
  return (
    <div className="flex items-center justify-between gap-6 px-5 py-4">
      <div className="flex items-start gap-3 min-w-0 flex-1">
        <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
          {icon}
        </span>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <p className="text-sm font-medium text-text-primary">{title}</p>
            {badge}
          </div>
          <p className="text-2xs text-text-muted mt-0.5 leading-relaxed">
            {description}
          </p>
        </div>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
}
