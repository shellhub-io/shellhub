import { useLayoutEffect, useRef } from "react";
import {
  useForm,
  type DefaultValues,
  type FieldValues,
  type Resolver,
  type UseFormProps,
  type UseFormReturn,
} from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import type { ZodType } from "zod";

/**
 * Standardized `useForm` for drawer/dialog forms. Every form in the app uses
 * this so validation wiring and the reset lifecycle are identical:
 *
 * - validation is always `zodResolver(schema)` (schemas are the single source
 *   of truth; form types come from `z.infer`);
 * - `mode: "onChange"` so the submit button reflects validity live;
 * - the form resets to `defaultValues` on every closed → open transition, so a
 *   reopened drawer never shows stale input.
 *
 * `defaultValues` may be rebuilt on each render (e.g. `buildXDefaults(entity)`);
 * only the latest value is used at reset time, so identity churn is harmless.
 */
export function useDrawerForm<TSchema extends ZodType<FieldValues>>(
  open: boolean,
  schema: TSchema,
  defaultValues: TSchema["_output"],
  options?: Omit<
    UseFormProps<TSchema["_output"]>,
    "resolver" | "mode" | "defaultValues"
  >,
): UseFormReturn<TSchema["_output"]> {
  type TValues = TSchema["_output"];

  const schemaRef = useRef(schema);

  useLayoutEffect(() => {
    schemaRef.current = schema;
  });

  const form = useForm<TValues>({
    mode: "onChange",
    resolver: ((values, context, opts) =>
      zodResolver(schemaRef.current)(
        values,
        context,
        opts,
      )) as Resolver<TValues>,
    defaultValues: defaultValues as DefaultValues<TValues>,
    ...options,
  });

  const { reset } = form;
  const wasOpen = useRef(open);

  useLayoutEffect(() => {
    if (open && !wasOpen.current) {
      wasOpen.current = true;
      reset(defaultValues);
    } else if (!open && wasOpen.current) {
      wasOpen.current = false;
    }
  }, [open, defaultValues, reset]);

  return form;
}
