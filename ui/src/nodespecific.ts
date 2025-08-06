import { Buffer } from "buffer";
import Process from "process";

const { env } = process;
globalThis.process = { ...Process, env };
globalThis.Buffer = Buffer;
