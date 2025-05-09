import { inject } from "vue";
import { SnackbarInjectionKey, type ISnackbarPlugin } from "@/plugins/snackbar";

const useSnackbar = (): ISnackbarPlugin => inject(SnackbarInjectionKey) as ISnackbarPlugin;

export default useSnackbar;
