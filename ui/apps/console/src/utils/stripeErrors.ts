/**
 * Map Stripe client-side error codes to user-friendly messages.
 * Codes come from `StripeError.code` returned by `stripe.createPaymentMethod`
 * and other Stripe.js surfaces.
 *
 * Reference: https://stripe.com/docs/error-codes
 */
const STRIPE_ERROR_MESSAGES: Record<string, string> = {
  card_declined: "Your card was declined. Please try a different payment method.",
  expired_card: "Your card has expired. Please use a different card.",
  incorrect_cvc: "Your card's security code (CVC) is incorrect.",
  incorrect_number: "Your card number is incorrect.",
  invalid_cvc: "Your card's security code (CVC) is invalid.",
  invalid_expiry_month: "Your card's expiration month is invalid.",
  invalid_expiry_year: "Your card's expiration year is invalid.",
  invalid_number: "Your card number is invalid.",
  processing_error: "An error occurred while processing your card. Please try again.",
  insufficient_funds: "Your card has insufficient funds.",
  lost_card: "Your card has been reported as lost. Please use a different payment method.",
  stolen_card: "Your card has been reported as stolen. Please use a different payment method.",
  authentication_required:
    "Your card requires authentication. Please complete the verification and try again.",
};

export function stripeErrorMessage(code: string | undefined, fallback?: string): string {
  if (code && STRIPE_ERROR_MESSAGES[code]) return STRIPE_ERROR_MESSAGES[code];
  return fallback ?? "Unable to process your card. Please try again.";
}
