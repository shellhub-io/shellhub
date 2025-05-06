import { inject } from "vue";
import { InjectionKey, type ISnackbarPlugin } from "@/plugins/snackbar";

const useSnackbar = (): ISnackbarPlugin => inject(InjectionKey) as ISnackbarPlugin;

export default useSnackbar;
