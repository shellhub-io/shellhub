import { adminApi } from "../../api/http";

const fetchDevices = (
  page: number,
  perPage: number,
  search: string,
  sortStatusField: string,
  sortStatusString: "asc" | "desc" | undefined,
) => {
  if (sortStatusField && sortStatusString) {
    return adminApi.getDevicesAdmin(
      search,
      page,
      perPage,
      undefined,
      sortStatusField,
      sortStatusString,
    );
  }

  return adminApi.getDevicesAdmin(search, page, perPage);
};

const getDevice = (uid: string) => adminApi.getDeviceAdmin(uid);

export { fetchDevices, getDevice };
