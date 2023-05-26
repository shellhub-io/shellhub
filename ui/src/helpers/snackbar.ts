import { inject } from "vue";
import { PluginInterface, InjectionKey } from "@/plugins/snackbar";

function useSnackbar(): PluginInterface {
  return inject(InjectionKey) as PluginInterface;
}

export default useSnackbar;
