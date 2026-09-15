import { down } from "./stack";

export default function globalTeardown() {
  down();
}
