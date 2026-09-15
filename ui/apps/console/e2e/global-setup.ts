import { seedInstance } from "./seed";
import { baseURL, up } from "./stack";

export default async function globalSetup() {
  up();
  await seedInstance(baseURL);
}
