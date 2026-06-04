declare module "asciinema-player" {
  export interface AsciinemaPlayerOptions {
    fit?: "width" | "height" | "both" | false;
    controls?: boolean;
    speed?: number;
    startAt?: number;
  }

  export interface AsciinemaPlayer {
    play(): void;
    pause(): void;
    seek(pos: number | string): Promise<void>;
    getCurrentTime(): Promise<number>;
    getDuration(): Promise<number | null>;
    dispose(): void;
    addEventListener(event: "playing" | "ended", handler: () => void): void;
    removeEventListener(event: "playing" | "ended", handler: () => void): void;
  }

  export function create(
    source: { data: string } | { url: string },
    container: HTMLElement,
    options?: AsciinemaPlayerOptions,
  ): AsciinemaPlayer;
}
