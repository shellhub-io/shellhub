import { create } from "zustand";

interface CreateNamespaceState {
  open: boolean;
  openDialog: () => void;
  closeDialog: () => void;
}

/**
 * Whether the create-namespace dialog is up. It is opened from the new-tab picker, which closes
 * and unmounts as it does, so the dialog lives in the layout and its state here.
 */
export const useCreateNamespaceStore = create<CreateNamespaceState>((set) => ({
  open: false,
  openDialog: () => set({ open: true }),
  closeDialog: () => set({ open: false }),
}));
