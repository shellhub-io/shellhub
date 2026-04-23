import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faCcAmex,
  faCcDinersClub,
  faCcDiscover,
  faCcJcb,
  faCcMastercard,
  faCcVisa,
} from "@fortawesome/free-brands-svg-icons";
import type { IconDefinition } from "@fortawesome/fontawesome-svg-core";
import { CreditCardIcon } from "@heroicons/react/24/outline";

const BRAND_ICONS: Record<string, IconDefinition> = {
  amex: faCcAmex,
  "diners-club": faCcDinersClub,
  discover: faCcDiscover,
  jcb: faCcJcb,
  mastercard: faCcMastercard,
  visa: faCcVisa,
};

interface BillingIconProps {
  brand: string;
  className?: string;
}

export default function BillingIcon({
  brand,
  className = "w-6 h-6",
}: BillingIconProps) {
  const icon = BRAND_ICONS[brand.toLowerCase()];
  if (!icon) {
    return <CreditCardIcon className={className} aria-hidden="true" />;
  }
  return (
    <FontAwesomeIcon
      icon={icon}
      className={className}
      aria-label={`${brand} card`}
    />
  );
}
