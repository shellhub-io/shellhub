import { useState } from "react";
import { apiErrorMessage } from "@/api/errors";

interface UseRowActionOptions<Row> {
  item: string;
  verb?: string;
  run: (row: Row) => Promise<unknown>;
}

/**
 * The confirmation state for an action on one row of a list: which row is targeted and why the
 * last attempt failed. confirm runs the action on the targeted row and does nothing when none is.
 * On success it clears the target, closing the dialog. On failure error reads
 * "Failed to {verb} {item}." followed by apiErrorMessage's text for the status, and the row stays
 * targeted so the dialog stays open. confirm never rejects: a failure reaches the caller only
 * through error, so awaiting confirm needs no catch.
 */
export function useRowAction<Row>({
  item,
  verb = "delete",
  run,
}: UseRowActionOptions<Row>) {
  const [target, setTarget] = useState<Row | null>(null);
  const [error, setError] = useState<string | null>(null);

  const open = (row: Row) => {
    setError(null);
    setTarget(row);
  };

  const close = () => {
    setError(null);
    setTarget(null);
  };

  const confirm = async () => {
    if (target === null) return;
    setError(null);
    try {
      await run(target);
      close();
    } catch (err) {
      setError(`Failed to ${verb} ${item}. ${apiErrorMessage(err)}`);
    }
  };

  return { target, open, close, confirm, error };
}
