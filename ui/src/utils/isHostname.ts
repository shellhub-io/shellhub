import { Filter, HostnameFilter } from "@/interfaces/IFilter";

const isHostname = (filter: Filter): filter is HostnameFilter => "hostname" in filter;
export default isHostname;
