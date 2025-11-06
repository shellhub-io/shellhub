declare module "asciinema-player" {
  export interface PlayerOptions {
    fit?: "width" | "height" | "both" | false;
    controls?: boolean | "auto";
    autoPlay?: boolean;
    loop?: boolean | number;
    startAt?: number | string;
    speed?: number;
    idleTimeLimit?: number;
    theme?: string;
    poster?: string;
    terminalFontSize?: string;
    terminalFontFamily?: string;
    terminalLineHeight?: number;
    preload?: boolean;
  }

  export interface AsciinemaPlayer {
    play(): void;
    pause(): void;
    seek(time: number): void;
    getCurrentTime(): Promise<number>;
    getDuration(): Promise<number>;
    addEventListener(event: string, handler: () => void): void;
    removeEventListener(event: string, handler: () => void): void;
    dispose(): void;
  }

  export function create(
    src: { data: string } | string,
    element: HTMLElement | null,
    opts?: PlayerOptions,
  ): AsciinemaPlayer;
}
