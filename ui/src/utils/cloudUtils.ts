import { computed } from "vue";
import { envVariables } from "../envVariables";

const isCloudEnvironment = () => computed(() => envVariables.isCloud);

export default isCloudEnvironment;
