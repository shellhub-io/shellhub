import {
  createContext,
  useContext,
  useEffect,
  useState,
  useRef,
  cloneElement,
  isValidElement,
  type ReactNode,
  type ReactElement,
  type ComponentPropsWithoutRef,
  type InputHTMLAttributes,
} from "react";
import {
  useFloating,
  useInteractions,
  useClick,
  useDismiss,
  useRole,
  useListNavigation,
  useTypeahead,
  useListItem,
  useMergeRefs,
  autoUpdate,
  offset,
  flip,
  shift,
  size,
  FloatingPortal,
  FloatingFocusManager,
  FloatingList,
  type Placement,
  type OpenChangeReason,
} from "@floating-ui/react";
import { cn } from "./cn";

type DropdownMode = "menu" | "combobox" | "content";

interface DropdownContextValue {
  open: boolean;
  setOpen: (open: boolean, event?: Event, reason?: OpenChangeReason) => void;
  mode: DropdownMode;
  portal: boolean;
  activeIndex: number | null;
  getItemProps: ReturnType<typeof useInteractions>["getItemProps"];
  getReferenceProps: ReturnType<typeof useInteractions>["getReferenceProps"];
  getFloatingProps: ReturnType<typeof useInteractions>["getFloatingProps"];
  refs: ReturnType<typeof useFloating>["refs"];
  floatingStyles: ReturnType<typeof useFloating>["floatingStyles"];
  context: ReturnType<typeof useFloating>["context"];
  elementsRef: React.RefObject<(HTMLElement | null)[]>;
  labelsRef: React.RefObject<(string | null)[]>;
  selectCallbacksRef: React.RefObject<
    Map<number, { onSelect: () => void; closeOnSelect: boolean }>
  >;
  setActiveIndex: (index: number | null) => void;
}

const DropdownContext = createContext<DropdownContextValue | null>(null);

function useDropdownContext() {
  const ctx = useContext(DropdownContext);
  if (!ctx)
    throw new Error(
      "Dropdown compound components must be used within <Dropdown>",
    );
  return ctx;
}

interface DropdownProps {
  children: ReactNode;
  placement?: Placement;
  portal?: boolean;
  mode?: DropdownMode;
  open?: boolean;
  onOpenChange?: (
    open: boolean,
    event?: Event,
    reason?: OpenChangeReason,
  ) => void;
}

function DropdownRoot({
  children,
  placement = "bottom-start",
  portal = false,
  mode = "menu",
  open: controlledOpen,
  onOpenChange,
}: DropdownProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const isControlled = controlledOpen !== undefined;
  const open = isControlled ? controlledOpen : uncontrolledOpen;
  const setOpen = (v: boolean, event?: Event, reason?: OpenChangeReason) => {
    if (!isControlled) setUncontrolledOpen(v);
    onOpenChange?.(v, event, reason);
  };

  const elementsRef = useRef<(HTMLElement | null)[]>([]);
  const labelsRef = useRef<(string | null)[]>([]);
  const selectCallbacksRef = useRef<
    Map<number, { onSelect: () => void; closeOnSelect: boolean }>
  >(new Map());
  const [activeIndex, setActiveIndex] = useState<number | null>(null);

  const { refs, floatingStyles, context } = useFloating({
    placement,
    open,
    onOpenChange: setOpen,
    whileElementsMounted: autoUpdate,
    middleware: [
      offset(4),
      flip(),
      shift({ padding: 8 }),
      size({
        apply({ availableHeight, rects, elements }) {
          Object.assign(elements.floating.style, {
            minWidth: `${rects.reference.width}px`,
            maxHeight: `${Math.max(100, availableHeight - 8)}px`,
          });
        },
      }),
    ],
  });

  const isItemMode = mode === "menu" || mode === "combobox";

  const click = useClick(context, {
    enabled: mode !== "combobox",
  });

  const dismiss = useDismiss(context);

  const role = useRole(context, {
    role: mode === "menu" ? "menu" : mode === "combobox" ? "listbox" : "dialog",
  });

  const listNav = useListNavigation(context, {
    listRef: elementsRef,
    activeIndex,
    onNavigate: setActiveIndex,
    enabled: isItemMode,
    loop: true,
    virtual: mode === "combobox",
    focusItemOnOpen: mode === "content" ? false : "auto",
  });

  const typeahead = useTypeahead(context, {
    listRef: labelsRef,
    activeIndex,
    onMatch: setActiveIndex,
    enabled: mode === "menu",
  });

  const { getReferenceProps, getFloatingProps, getItemProps } = useInteractions(
    [click, dismiss, role, listNav, typeahead],
  );

  return (
    <DropdownContext.Provider
      value={{
        open,
        setOpen,
        mode,
        portal,
        activeIndex,
        getItemProps,
        getReferenceProps,
        getFloatingProps,
        refs,
        floatingStyles,
        context,
        elementsRef,
        labelsRef,
        selectCallbacksRef,
        setActiveIndex,
      }}
    >
      {children}
    </DropdownContext.Provider>
  );
}

function Trigger({ children }: { children: ReactElement }) {
  const { refs, getReferenceProps } = useDropdownContext();

  const childRef = isValidElement(children)
    ? (children as ReactElement & { ref?: React.Ref<HTMLElement> }).ref
    : undefined;
  // eslint-disable-next-line @typescript-eslint/unbound-method -- Floating UI callback ref setter, does not use `this`
  const mergedRef = useMergeRefs([refs.setReference, childRef].filter(Boolean));

  if (!isValidElement(children)) return null;

  return cloneElement(children, {
    ref: mergedRef,
    ...getReferenceProps(children.props as Record<string, unknown>),
  } as Record<string, unknown>);
}

function Anchor({
  children,
  className,
  ...props
}: ComponentPropsWithoutRef<"div">) {
  const { refs } = useDropdownContext();

  return (
    <div
      // eslint-disable-next-line react-hooks/refs, @typescript-eslint/unbound-method -- Floating UI callback ref setter
      ref={refs.setPositionReference}
      className={className}
      {...props}
    >
      {children}
    </div>
  );
}

type InputProps = Omit<InputHTMLAttributes<HTMLInputElement>, "onChange"> & {
  onChange?: (value: string) => void;
};

function Input({ onChange, ...props }: InputProps) {
  const {
    refs,
    getReferenceProps,
    open,
    setOpen,
    activeIndex,
    setActiveIndex,
    selectCallbacksRef,
  } = useDropdownContext();
  const blurTimeout = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => () => clearTimeout(blurTimeout.current), []);

  // eslint-disable-next-line react-hooks/refs -- Floating UI callback ref setter
  const inputProps = getReferenceProps({
    ref: refs.setReference, // eslint-disable-line @typescript-eslint/unbound-method
    ...props,
    onChange: (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange?.(e.target.value);
      if (!open) setOpen(true);
      setActiveIndex(0);
    },
    onFocus: (e: React.FocusEvent<HTMLInputElement>) => {
      clearTimeout(blurTimeout.current);
      if (!open) setOpen(true);
      props.onFocus?.(e);
    },
    onBlur: (e: React.FocusEvent<HTMLInputElement>) => {
      const floatingEl = refs.floating.current;
      if (floatingEl?.contains(e.relatedTarget)) return;
      blurTimeout.current = setTimeout(() => setOpen(false), 150);
      props.onBlur?.(e);
    },
    onKeyDown: (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter" && activeIndex !== null) {
        const entry = selectCallbacksRef.current.get(activeIndex);
        if (entry) {
          e.preventDefault();
          entry.onSelect();
          if (entry.closeOnSelect !== false) setOpen(false);
        }
      }
      props.onKeyDown?.(e);
    },
  });

  return <input aria-autocomplete="list" {...inputProps} />;
}

interface PanelProps extends ComponentPropsWithoutRef<"div"> {
  "aria-label"?: string;
}

function Panel({ children, className, ...props }: PanelProps) {
  const {
    open,
    mode,
    portal,
    refs,
    floatingStyles,
    getFloatingProps,
    context,
    elementsRef,
    labelsRef,
  } = useDropdownContext();

  if (!open) return null;

  const isItemMode = mode === "menu" || mode === "combobox";

  const panel = (
    <div
      // eslint-disable-next-line react-hooks/refs, @typescript-eslint/unbound-method -- Floating UI callback ref setter
      ref={refs.setFloating}
      style={floatingStyles}
      className={cn(
        "bg-surface border border-border rounded-xl shadow-2xl animate-fade-in overflow-y-auto z-dropdown",
        className,
      )}
      {...getFloatingProps(props)}
    >
      {isItemMode ? (
        <FloatingList elementsRef={elementsRef} labelsRef={labelsRef}>
          {children}
        </FloatingList>
      ) : (
        children
      )}
    </div>
  );

  const content =
    mode === "combobox" ? (
      panel
    ) : (
      <FloatingFocusManager context={context} initialFocus={0} returnFocus>
        {panel}
      </FloatingFocusManager>
    );

  return portal ? <FloatingPortal>{content}</FloatingPortal> : content;
}

interface ItemProps extends ComponentPropsWithoutRef<"div"> {
  label?: string;
  onSelect?: () => void;
  disabled?: boolean;
  variant?: "default" | "danger" | "create";
  closeOnSelect?: boolean;
}

function Item({
  children,
  label,
  onSelect,
  disabled = false,
  variant = "default",
  closeOnSelect = true,
  className,
  ...props
}: ItemProps) {
  const { activeIndex, getItemProps, mode, setOpen, selectCallbacksRef } =
    useDropdownContext();
  const { ref, index } = useListItem({ label: disabled ? null : label });
  const isActive = activeIndex === index;

  const itemRole = mode === "combobox" ? "option" : "menuitem";

  useEffect(() => {
    const map = selectCallbacksRef.current;
    if (onSelect && !disabled) {
      map.set(index, { onSelect, closeOnSelect });
    }
    return () => {
      map.delete(index);
    };
  }, [index, onSelect, disabled, closeOnSelect, selectCallbacksRef]);

  return (
    <div
      ref={ref}
      role={itemRole}
      tabIndex={mode === "combobox" ? -1 : isActive ? 0 : -1}
      aria-disabled={disabled || undefined}
      aria-selected={mode === "combobox" ? isActive : undefined}
      className={cn(
        "flex items-center px-2.5 py-1.5 text-xs rounded-md outline-none transition-colors cursor-pointer",
        isActive && "bg-hover-medium",
        disabled && "opacity-dim pointer-events-none",
        variant === "danger"
          ? "text-accent-red hover:bg-accent-red/10"
          : variant === "create"
            ? "text-accent-green hover:bg-hover-medium"
            : "text-text-secondary hover:bg-hover-medium",
        className,
      )}
      {...getItemProps({
        ...props,
        onClick(e: React.MouseEvent<HTMLDivElement>) {
          if (disabled) return;
          onSelect?.();
          if (closeOnSelect) setOpen(false);
          props.onClick?.(e);
        },
        onKeyDown(e: React.KeyboardEvent<HTMLDivElement>) {
          if (disabled) return;
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onSelect?.();
            if (closeOnSelect) setOpen(false);
          }
          props.onKeyDown?.(e);
        },
      })}
    >
      {children}
    </div>
  );
}

export const Dropdown = Object.assign(DropdownRoot, {
  Trigger,
  Anchor,
  Input,
  Panel,
  Item,
});
