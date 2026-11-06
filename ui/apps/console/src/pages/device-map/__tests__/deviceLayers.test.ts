import { describe, expect, it } from "vitest";
import {
  createPropertyExpression,
  latest,
  type StylePropertySpecification,
} from "@maplibre/maplibre-gl-style-spec";
import type { ExpressionSpecification } from "maplibre-gl";
import { mockMapPalette } from "@/tests/factories";
import { color, label, radius } from "../deviceLayers";

const palette = mockMapPalette();

type Property = "circle-radius" | "circle-color" | "text-field";

function drawn(
  expression: ExpressionSpecification,
  property: Property,
  properties: { devices: number; online: number },
): string {
  const specs: Record<Property, unknown> = {
    "circle-radius": latest.paint_circle["circle-radius"],
    "circle-color": latest.paint_circle["circle-color"],
    "text-field": latest.layout_symbol["text-field"],
  };
  const parsed = createPropertyExpression(
    expression,
    `layers[0].${property}`,
    specs[property] as StylePropertySpecification,
  );
  if (parsed.result === "error") {
    throw new Error(parsed.value.map((error) => error.message).join("; "));
  }

  return String(parsed.value.evaluate({ zoom: 0 }, { type: 1, properties }));
}

describe("radius", () => {
  it.each([
    { devices: 1, total: 10, px: 20.6491106 },
    { devices: 10, total: 10, px: 48 },
    { devices: 1, total: 10_000, px: 8.4 },
    { devices: 0, total: 0, px: 8 },
  ])("draws $devices of $total devices at $px px", ({ devices, total, px }) => {
    expect(
      Number(drawn(radius("all", total), "circle-radius", { devices, online: 0 })),
    ).toBeCloseTo(px, 5);
  });

  it("sizes a circle by the devices the mode counts", () => {
    const counts = { devices: 10, online: 1 };

    expect(Number(drawn(radius("online", 10), "circle-radius", counts))).toBeCloseTo(
      20.6491106,
      5,
    );
    expect(Number(drawn(radius("offline", 10), "circle-radius", counts))).toBeCloseTo(
      45.9473319,
      5,
    );
  });
});

describe("color", () => {
  it.each([
    { online: 0, rgba: "rgba(200,0,0,1)" },
    { online: 5, rgba: "rgba(200,0,0,1)" },
    { online: 8, rgba: "rgba(200,200,0,1)" },
    { online: 9, rgba: "rgba(100,200,0,1)" },
    { online: 10, rgba: "rgba(0,200,0,1)" },
  ])("colours $online of 10 online as $rgba", ({ online, rgba }) => {
    const expression = color("all", palette) as ExpressionSpecification;

    expect(drawn(expression, "circle-color", { devices: 10, online })).toBe(rgba);
  });

  it("draws one colour per state under online and offline", () => {
    expect(color("online", palette)).toBe(palette.online);
    expect(color("offline", palette)).toBe(palette.muted);
  });
});

describe("label", () => {
  it.each([
    { devices: 1, text: "" },
    { devices: 2, text: "2" },
    { devices: 999, text: "999" },
    { devices: 1000, text: "1k" },
    { devices: 13_310, text: "13.3k" },
  ])("labels $devices devices as '$text'", ({ devices, text }) => {
    expect(drawn(label("all"), "text-field", { devices, online: 0 })).toBe(text);
  });
});
