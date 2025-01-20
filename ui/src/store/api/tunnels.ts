import { tunnelApi } from "../../api/http";

const getTunnels = (uid: string) => tunnelApi.listTunnels(uid);

const createTunnel = (uid: string, host: string, port: number, ttl: number) => tunnelApi.createTunnel(
  uid,
  {
    host,
    port,
    ttl,
  },
);

const deleteTunnel = (uid: string, address: string) => tunnelApi.deleteTunnel(uid, address);

export { getTunnels, createTunnel, deleteTunnel };
