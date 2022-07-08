import { Buffer } from "buffer";
import Process from "process";

globalThis.process = Process;
globalThis.Buffer = Buffer;