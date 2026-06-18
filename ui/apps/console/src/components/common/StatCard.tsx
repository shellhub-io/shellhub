import { ReactNode } from "react";
import { Link } from "react-router-dom";
import { Card } from "@shellhub/design-system/primitives";

interface StatCardBaseProps {
  icon: ReactNode;
  title: string;
  value: number | string;
  linkLabel: string;
  accent?: string;
}

interface StatCardLinkProps extends StatCardBaseProps {
  linkTo: string;
  onClick?: never;
}

interface StatCardButtonProps extends StatCardBaseProps {
  onClick: () => void;
  linkTo?: never;
}

type StatCardProps = StatCardLinkProps | StatCardButtonProps;

export default function StatCard(props: StatCardProps) {
  const { icon, title, value, linkLabel, accent } = props;

  return (
    <Card className="rounded-lg p-6 flex flex-col items-center text-center group hover:border-primary/30 transition-all duration-300">
      <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary mb-5">
        {icon}
      </div>

      <p className="text-2xs font-mono font-medium uppercase tracking-label text-text-muted mb-2">
        {title}
      </p>

      <p
        className={`text-4xl font-mono font-bold mb-5 tabular-nums ${accent ?? "text-text-primary"}`}
      >
        {value}
      </p>

      {props.onClick ? (
        <button
          type="button"
          onClick={props.onClick}
          className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
        >
          {linkLabel} &rarr;
        </button>
      ) : (
        <Link
          to={props.linkTo}
          className="text-xs font-medium text-primary hover:text-primary-400 transition-colors"
        >
          {linkLabel} &rarr;
        </Link>
      )}
    </Card>
  );
}
