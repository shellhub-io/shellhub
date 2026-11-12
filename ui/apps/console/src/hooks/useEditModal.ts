import { useState } from "react";

/**
 * useEditModal holds the state of a modal that both creates and edits a row. openNew opens it
 * empty, openEdit opens it on a row, and target is that row, or null while creating. close keeps
 * target until the next openNew or openEdit, so read it only while open.
 */
export function useEditModal<Row>() {
  const [isOpen, setIsOpen] = useState(false);
  const [target, setTarget] = useState<Row | null>(null);

  return {
    isOpen,
    target,
    openNew: () => {
      setTarget(null);
      setIsOpen(true);
    },
    openEdit: (row: Row) => {
      setTarget(row);
      setIsOpen(true);
    },
    close: () => setIsOpen(false),
  };
}
